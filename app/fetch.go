package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/mwinters0/hnjobs/db"
	"github.com/mwinters0/hnjobs/hn"
	"github.com/mwinters0/hnjobs/scoring"
	"html"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var existingJobs map[int]*db.Job // key is job ID
var numNewJobsFetched atomic.Int32
var numUpdatedJobsFetched atomic.Int32
var numCommentsFetched atomic.Int32

type UpdateType = int

const (
	UpdateTypeGeneric  UpdateType = iota
	UpdateTypeNewStory            // value is new story id
	UpdateTypeNonFatalErr
	UpdateTypeFatal
	UpdateTypeBadComment
	UpdateTypeJobFetched
	UpdateTypeDone // value is newJobs + updatedJobs
)

type FetchStatusUpdate struct {
	UpdateType UpdateType
	Message    string
	Value      int // either new job score or number of jobs fetched on completion
	Error      error
}

const WhoIsHiringString string = "Who is hiring?"

// todo? make this config-driven
const maxCompanyNameLength = 30 // is not the same const in app/browse

type FetchOptions struct {
	Context     context.Context
	Status      chan<- FetchStatusUpdate
	ModeForce   bool
	StoryID     int // fetch latest if 0
	TTLSec      int64
	MustContain string // typically "Who's Hiring"
}

func genericStatus(s string, c chan<- FetchStatusUpdate) {
	c <- FetchStatusUpdate{UpdateTypeGeneric, s, 0, nil}
}

func FetchAsync(fo FetchOptions) {
	var err error
	existingJobs = make(map[int]*db.Job) // cache for TTL checking
	numNewJobsFetched.Store(0)
	numUpdatedJobsFetched.Store(0)
	numCommentsFetched.Store(0)
	storyId := fo.StoryID // if we fetch latest then this val will change

	notifyCompletion := func(msg string, v int, e error, fatal bool) {
		// Just a single place to close() on completion
		if fatal {
			fo.Status <- FetchStatusUpdate{UpdateTypeFatal, msg, v, e}
		} else {
			fo.Status <- FetchStatusUpdate{UpdateTypeDone, msg, v, e}
		}
		close(fo.Status)
	}

	isNewStory := false
	var apiStory *hn.Story
	if storyId == 0 {
		// get latest
		genericStatus("Fetching job stories...", fo.Status)
		submissions, err := hn.FetchSubmissions(fo.Context, "whoishiring")
		if err != nil {
			notifyCompletion("Error fetching job stories", 0, err, true)
			return
		}
		genericStatus(fmt.Sprintf("Found %d job stories...", len(submissions)), fo.Status)
		if fo.MustContain != "" {
			genericStatus(fmt.Sprintf("Searching for most-recent '%s' story...", fo.MustContain), fo.Status)
			for i := range submissions {
				s, err := hn.FetchStory(fo.Context, submissions[i])
				if err != nil {
					notifyCompletion(fmt.Sprintf("Failed to retrieve job story %d from API.", submissions[i]), 0, err, true)
					return
				}
				if strings.Contains(s.Title, fo.MustContain) {
					apiStory = s
					break
				}
			}
			if apiStory == nil {
				notifyCompletion(fmt.Sprintf("Couldn't find a job story matching '%s'", fo.MustContain), 0, err, true)
				return
			}
		} else {
			apiStory, err = hn.FetchStory(fo.Context, submissions[0])
			if err != nil {
				notifyCompletion(fmt.Sprintf("Failed to retrieve job story %d from API.", submissions[0]), 0, err, true)
				return
			}
		}
		storyId = apiStory.Id
	} else {
		apiStory, err = hn.FetchStory(fo.Context, storyId)
		if err != nil {
			notifyCompletion("Failed to retrieve job story from API.", 0, err, true)
			return
		}
	}

	// get story comment IDs

	dbStory, err := db.GetStoryById(storyId)
	if err != nil && !errors.Is(err, db.ErrNoResults) {
		notifyCompletion("Failure checking DB for existing story.", 0, err, true)
		return
	}
	if errors.Is(err, db.ErrNoResults) {
		isNewStory = true
		fo.Status <- FetchStatusUpdate{
			UpdateTypeNewStory,
			fmt.Sprintf("Found NEW job story: \"%s\" (%d top-level comments)", apiStory.Title, len(apiStory.Kids)),
			apiStory.Id,
			nil,
		}
		// We don't want to set fetched_time in the DB until after we've completed the fetch
		apiStory.FetchedTime = 0
		apiStory.FetchedGoTime = time.Unix(0, 0)
		err = db.UpsertStory(apiStory)
		if err != nil {
			notifyCompletion("Failed to upsert story.", 0, err, true)
			return
		}
	} else {
		genericStatus(fmt.Sprintf("Most-recent job story (%d) was previously cached", storyId), fo.Status)
	}

	// if we're using story-based ttl, then we only have two modes: fetch all (ttl expired), or fetch new.
	// However, we want to defer updating the story fetch time until after fetch is complete.

	var commentIDsToFetch []int
	if isNewStory {
		commentIDsToFetch = apiStory.Kids
		genericStatus("Fetching all top-level comments...", fo.Status)
	} else {
		// When fetching a job, it might be an update of an existing job.  We then want to set unread while
		// preserving the rest of the user-created state.
		jobs, err := db.GetAllJobsByStoryId(storyId, db.OrderNone)
		if err != nil && errors.Is(err, db.ErrNoResults) {
			notifyCompletion("Failed to fetch existing jobs from the DB", 0, err, true)
			return
		}
		for _, job := range jobs {
			existingJobs[job.Id] = job
		}
		//decide what we're fetching
		if fo.ModeForce {
			// fetch all
			commentIDsToFetch = apiStory.Kids
			genericStatus("ModeForce-fetching all top-level comments...", fo.Status)
		} else {
			// check story TTL
			age := time.Now().UTC().Unix() - dbStory.FetchedTime
			if age > fo.TTLSec {
				// expired TTL - fetch all
				genericStatus(
					fmt.Sprintf("Story TTL is expired (age %d, TTL is %d) - fetching all", age, fo.TTLSec),
					fo.Status,
				)
				commentIDsToFetch = apiStory.Kids
			} else {
				// only fetch new
				genericStatus(
					fmt.Sprintf("Story TTL is not expired (age %d, TTL is %d) - only fetching new", age, fo.TTLSec),
					fo.Status,
				)
				for _, id := range apiStory.Kids {
					if _, ok := existingJobs[id]; !ok {
						commentIDsToFetch = append(commentIDsToFetch, id)
					}
				}
				existingJobs = make(map[int]*db.Job) // free this memory / speed up this search
			}
		}
	}

	// pipeline: produce comment IDs -> fetch comments -> score and store -> done
	//
	// producer
	commentIDs := make(chan int)
	produce := func() {
		for _, commentID := range commentIDsToFetch {
			select {
			case <-fo.Context.Done():
				close(commentIDs)
				return
			default:
				commentIDs <- commentID
			}
		}
		close(commentIDs)
	}
	// fetchers
	numWorkers := 3
	comments := make(chan *hn.Comment)
	workerUpdates := make(chan FetchStatusUpdate)
	fwg := sync.WaitGroup{}
	fwg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go commentFetcher(fo.Context, &fwg, commentIDs, comments, workerUpdates)
	}
	// processors
	pwg := sync.WaitGroup{}
	pwg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go commentProcessor(fo.Context, &pwg, comments, workerUpdates)
	}
	// done waiter
	workersDone := make(chan int)
	go func() {
		fwg.Wait()
		close(comments)
		pwg.Wait()
		close(workersDone)
	}()

	go produce()

	totalJobsFetched := func() int {
		return int(numNewJobsFetched.Load() + numUpdatedJobsFetched.Load())
	}
	// Wait for one of the following:
	// 1. cancellation
	// 2. all comments to be finished
	// 3. a fatal error from a worker (cancel all other workers)
	for {
		select {
		case <-fo.Context.Done(): //1
			notifyCompletion("Fetching cancelled", totalJobsFetched(), nil, false)
			return
		case <-workersDone: //2
			// update the fetched_time on the story for TTL
			apiStory.FetchedGoTime = time.Now()
			apiStory.FetchedTime = apiStory.FetchedGoTime.UTC().Unix()
			err = db.UpsertStory(apiStory)
			if err != nil {
				notifyCompletion("Failed to upsert story.", totalJobsFetched(), err, true)
				return
			}
			notifyCompletion(
				fmt.Sprintf(
					"Done. Fetched %d new jobs, %d updated jobs (%d comments).",
					numNewJobsFetched.Load(),
					numUpdatedJobsFetched.Load(),
					numCommentsFetched.Load(),
				),
				int(numNewJobsFetched.Load()+numUpdatedJobsFetched.Load()),
				nil,
				false,
			)
			return
		case wStatus := <-workerUpdates: //this channel should never close
			if wStatus.UpdateType == UpdateTypeFatal { //3
				notifyCompletion(wStatus.Message, wStatus.Value, wStatus.Error, true)
				return
			}
			fo.Status <- wStatus
		}
	}
}

func commentFetcher(ctx context.Context, wg *sync.WaitGroup, commentIDs <-chan int, comments chan<- *hn.Comment, status chan<- FetchStatusUpdate) {
	for {
		select {
		case <-ctx.Done():
			//cancelled
			wg.Done()
			return
		case i, ok := <-commentIDs:
			if !ok {
				//EOF
				wg.Done()
				return
			}
			c, err := hn.FetchComment(ctx, i)
			if err != nil {
				status <- FetchStatusUpdate{
					UpdateTypeNonFatalErr,
					fmt.Sprintf("Failed to fetch comment id %d from API! Ignoring.", i),
					0,
					err,
				}
				// TODO retry?  backoff?
				continue
			}
			numCommentsFetched.Add(1)
			if len(c.Text) == 0 {
				status <- FetchStatusUpdate{
					UpdateTypeBadComment,
					fmt.Sprintf("Got empty comment id %d from API, ignoring", i),
					0,
					errors.New("empty"),
				}
				continue
			}
			comments <- c
		}
	}
}

// commentProcessor converts a hn Comment to a job, scores it, and stores it in the DB.
func commentProcessor(ctx context.Context, wg *sync.WaitGroup, comments <-chan *hn.Comment, status chan<- FetchStatusUpdate) {
	for {
		select {
		case <-ctx.Done():
			//cancelled
			wg.Done()
			return
		case c, ok := <-comments:
			if !ok {
				//EOF
				wg.Done()
				return
			}
			if len(c.Text) == 0 {
				status <- FetchStatusUpdate{
					UpdateTypeBadComment,
					fmt.Sprintf("Bad comment (id %d): empty comment", c.Id),
					0,
					errors.New("len==0"),
				}
				continue
			}
			if !strings.Contains(c.Text, "|") {
				status <- FetchStatusUpdate{
					UpdateTypeBadComment,
					fmt.Sprintf("Bad comment (id %d): not a job comment", c.Id),
					0,
					errors.New("doesn't contain \"|\""),
				}
				continue
			}
			cname, err := getCompanyName(c.Text, maxCompanyNameLength)
			if err != nil {
				status <- FetchStatusUpdate{
					UpdateTypeBadComment,
					fmt.Sprintf("Bad comment (id %d): couldn't find company name", c.Id),
					0,
					err,
				}
				continue
			}
			job, err := newJobFromHNComment(c, cname)
			if err != nil {
				status <- FetchStatusUpdate{
					UpdateTypeNonFatalErr,
					fmt.Sprintf("Unable to process comment ID %d: %v", c.Id, err),
					0,
					err,
				}
				continue
			}
			job.FetchedTime = time.Now().UTC().Unix()
			score := scoring.ScoreDBComment(job)
			// check existing
			existingJob, found := existingJobs[c.Id]
			if found {
				// preserve user state
				job.Applied = existingJob.Applied
				job.Priority = existingJob.Priority
				job.Interested = existingJob.Interested
				if job.Text != existingJob.Text {
					// text changed since last time we saw it
					numUpdatedJobsFetched.Add(1)
					job.Read = false
				} else {
					job.Read = existingJob.Read
				}
			} else {
				numNewJobsFetched.Add(1)
			}
			err = db.UpsertJob(job)
			if err != nil {
				//fatal
				status <- FetchStatusUpdate{
					UpdateTypeFatal,
					fmt.Sprintf("Failed to upsert job into DB!"),
					0,
					err,
				}
				wg.Done()
				return
			}
			status <- FetchStatusUpdate{
				UpdateTypeJobFetched,
				fmt.Sprintf("New job (%d): [Score %d]", c.Id, score),
				score,
				nil,
			}
		}
	}
}

func newJobFromHNComment(c *hn.Comment, companyName string) (*db.Job, error) {
	job := &db.Job{
		Id:            c.Id,
		Parent:        c.Parent,
		Company:       companyName,
		Text:          c.Text,
		Time:          c.Time,
		GoTime:        c.GoTime,
		FetchedTime:   c.FetchedTime,
		FetchedGoTime: c.FetchedGoTime,
		Read:          false,
		Interested:    true,
		Priority:      false,
		Applied:       false,
	}
	return job, nil
}

func getCompanyName(text string, maxlen int) (string, error) {
	if len(text) == 0 {
		return "", errors.New("text len is 0")
	}
	text = html.UnescapeString(text)
	i := strings.IndexAny(text, "(|<\n")
	if i < 1 {
		// Note: some jokers put a URL as the company name, causing this.  We could special-case that, but maybe
		// just don't work for jokers?  You're welcome.
		return "", errors.New("no text before first delimiter")
	}
	if i > maxlen {
		i = maxlen
	}
	return strings.Clone(strings.TrimSpace(text[:i])), nil
}
