package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	//_ "github.com/ncruces/go-sqlite3/driver" //sqlite3
	//_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/mwinters0/hnjobs/hn"
	_ "modernc.org/sqlite" //sqlite
	"strconv"
	"sync"
)

var store = sqlStore{}

type sqlStore struct {
	db         *sql.DB
	jobUpsert  *sql.Stmt
	writeMutex sync.Mutex //needed? I was trying different sqlite drivers and not sure all were threadsafe
}

// scannableRow is either a sql.Row or sql.Rows to facilitate unmarshalling, because those two are not related???
type scannableRow interface {
	Scan(dest ...interface{}) error
}

var ErrNoResults = errors.New("no results") // hide the sql package

type JobOrder int

const (
	OrderNone JobOrder = iota
	OrderScoreDesc
	OrderScoreAsc
	OrderTimeNewestFirst
	OrderTimeOldestFirst
)

func OpenDB(filePath string) error {
	var err error
	store.db, err = sql.Open("sqlite", "file:"+filePath)
	if err != nil {
		return fmt.Errorf("error opening DB: %v", err)
	}
	return nil
}

// === table: stories

func GetStoryById(id int) (*hn.Story, error) {
	storyRow := store.db.QueryRow(storySelect+"WHERE Id = ?", strconv.Itoa(id))
	story, err := unmarshalStoryRow(storyRow)
	if errors.Is(err, sql.ErrNoRows) {
		return story, ErrNoResults
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving story: %v", err)
	}
	return story, nil
}

func GetAllStories() ([]*hn.Story, error) {
	stories := []*hn.Story{}
	rows, err := store.db.Query(storySelect + "ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("error retrieving stories: %v", err)
	}
	for rows.Next() {
		story, err := unmarshalStoryRow(rows)
		if errors.Is(err, sql.ErrNoRows) {
			return stories, ErrNoResults
		} else if err != nil {
			return stories, fmt.Errorf("couldn't unmarshal story: %v", err)
		}
		stories = append(stories, story)
	}
	return stories, nil
}

func GetLatestStory() (*hn.Story, error) {
	storyRow := store.db.QueryRow(storySelect + "ORDER BY id DESC LIMIT 1")
	story, err := unmarshalStoryRow(storyRow)
	if errors.Is(err, sql.ErrNoRows) {
		return story, ErrNoResults
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving latest story: %v", err)
	}
	return story, nil
}

func DeleteStoryAndJobsByStoryID(id int) error {
	store.writeMutex.Lock()
	_, err := store.db.Exec(`DELETE FROM hnstories WHERE id = ?`, strconv.Itoa(id))
	if err != nil {
		return fmt.Errorf("error deleting story: %v", err)
	}
	_, err = store.db.Exec(`DELETE FROM hnjobs WHERE parent = ?`, strconv.Itoa(id))
	if err != nil {
		return fmt.Errorf("error deleting jobs: %v", err)
	}
	store.writeMutex.Unlock()
	return nil
}

const storySelect = "SELECT id, kids, time, title, fetched_time FROM hnstories "

func unmarshalStoryRow(row scannableRow) (*hn.Story, error) {
	s := hn.Story{}
	kids := sql.NullString{}
	err := row.Scan(&s.Id, &kids, &s.Time, &s.Title, &s.FetchedTime)
	if err != nil {
		return &hn.Story{}, err
	}
	s.GoTime = time.Unix(s.Time, 0)
	s.FetchedGoTime = time.Unix(s.FetchedTime, 0)
	if kids.Valid {
		err = json.Unmarshal([]byte(kids.String), &s.Kids)
		if err != nil {
			log.Fatal(err)
		}
	}
	return &s, nil
}

func UpsertStory(s *hn.Story) error {
	kids, err := json.Marshal(s.Kids)
	if err != nil {
		log.Fatal(err)
	}
	store.writeMutex.Lock()
	_, err = store.db.Exec(
		`INSERT INTO hnstories (id, kids, time, title, fetched_time) VALUES (?, ?, ?, ?, ?)
	ON CONFLICT (id) DO UPDATE SET
	kids = excluded.kids, time = excluded.time, title = excluded.title, fetched_time = excluded.fetched_time`,
		s.Id, nullableString(kids), s.Time, s.Title, s.FetchedTime,
	)
	store.writeMutex.Unlock()
	if err != nil {
		return fmt.Errorf("upsert failed: %v", err)
	}
	return nil
}

// === table: hnjobs

type Job struct {
	Id             int
	Parent         int
	Company        string
	Text           string
	Time           int64
	GoTime         time.Time `json:"-"`
	FetchedTime    int64
	FetchedGoTime  time.Time `json:"-"`
	ReviewedTime   int64
	ReviewedGoTime time.Time `json:"-"`
	Why            []string
	WhyNot         []string
	Score          int
	Read           bool
	Interested     bool
	Priority       bool
	Applied        bool
}

func UpsertJob(job *Job) error {
	if store.jobUpsert == nil {
		store.jobUpsert, _ = store.db.Prepare(
			`INSERT INTO hnjobs (
			id, parent, company, text, time, fetched_time,
			reviewed_time, score, why, why_not,
			read, interested, priority, applied
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
			company=excluded.company, text=excluded.text, time=excluded.time, fetched_time=excluded.fetched_time,
			reviewed_time=excluded.reviewed_time, score=excluded.score, why=excluded.why, why_not=excluded.why_not,
			read=excluded.read, interested=excluded.interested, priority=excluded.priority, applied=excluded.applied 
			`,
		)
	}
	why, err := json.Marshal(job.Why)
	if err != nil {
		log.Fatal(err)
	}
	whyNot, err := json.Marshal(job.WhyNot)
	if err != nil {
		log.Fatal(err)
	}
	store.writeMutex.Lock()
	_, err = store.jobUpsert.Exec(
		job.Id, job.Parent, job.Company, job.Text, job.Time, job.FetchedTime,
		job.ReviewedTime, job.Score, nullableString(why), nullableString(whyNot),
		job.Read, job.Interested, job.Priority, job.Applied,
	)
	store.writeMutex.Unlock()
	if err != nil {
		return fmt.Errorf("upsert failed: %v", err)
	}
	return nil
}

func GetAllJobsByStoryId(id int, co JobOrder) ([]*Job, error) {
	var jobs []*Job
	orderBy := ""
	switch co {
	case OrderNone:
		orderBy = "id"
	case OrderScoreDesc:
		orderBy = "score DESC"
	case OrderScoreAsc:
		orderBy = "score ASC"
	case OrderTimeNewestFirst:
		orderBy = "time DESC"
	case OrderTimeOldestFirst:
		orderBy = "time ASC"
	default:
		panic(errors.New("unhandled JobOrder"))
	}
	rows, err := store.db.Query(jobSelect+"WHERE parent = ? ORDER BY "+orderBy, id)
	if err != nil {
		return jobs, fmt.Errorf("couldn't query: %v", err)
	}
	for rows.Next() {
		job, err := unmarshalJobRow(rows)
		if errors.Is(err, sql.ErrNoRows) {
			return jobs, ErrNoResults
		} else if err != nil {
			return jobs, fmt.Errorf("couldn't unmarshal job: %v", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func GetAllJobIDsByStoryID(storyID int) *[]int {
	var i []int
	// ordered to match the results from HN API
	jobRows, err := store.db.Query("SELECT id FROM hnjobs WHERE parent = ? ORDER BY id ASC", storyID)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to retrieve job ids: %v", err))
	}
	for jobRows.Next() {
		var id int
		err = jobRows.Scan(&id)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to parse existing job storyID: %v", err))
		}
		i = append(i, id)
	}
	return &i
}

const jobSelect = `SELECT id, parent, company, text, time, fetched_time,
	reviewed_time, why, why_not, score,
	read, interested, priority, applied FROM hnjobs
`

func unmarshalJobRow(row scannableRow) (*Job, error) {
	job := Job{}
	reviewedTime := sql.NullInt64{}
	why := sql.NullString{}
	whyNot := sql.NullString{}
	err := row.Scan(
		&job.Id, &job.Parent, &job.Company, &job.Text, &job.Time, &job.FetchedTime,
		&reviewedTime, &why, &whyNot, &job.Score,
		&job.Read, &job.Interested, &job.Priority, &job.Applied,
	)
	if err != nil {
		return &Job{}, err
	}
	if reviewedTime.Valid {
		job.ReviewedTime = reviewedTime.Int64
		job.ReviewedGoTime = time.Unix(job.ReviewedTime, 0)
	}
	if why.Valid {
		err = json.Unmarshal([]byte(why.String), &job.Why)
		if err != nil {
			log.Fatal(err)
		}
	}
	if whyNot.Valid {
		err = json.Unmarshal([]byte(whyNot.String), &job.WhyNot)
		if err != nil {
			log.Fatal(err)
		}
	}
	job.GoTime = time.Unix(job.Time, 0)
	job.FetchedGoTime = time.Unix(job.FetchedTime, 0)
	return &job, nil
}

// === util

// avoid inserting "null" for empty strings
func nullableString(in []byte) sql.NullString {
	s := string(in)
	out := sql.NullString{}
	if s != "null" {
		out.Valid = true
		out.String = s
	}
	return out
}
