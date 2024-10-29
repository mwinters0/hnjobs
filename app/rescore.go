package app

import (
	"errors"
	"hnjobs/db"
	"hnjobs/scoring"
)

func ReScore(storyID int) (int, error) {
	dbcs, err := db.GetAllJobsByStoryId(storyID, db.OrderNone)
	if err != nil {
		return 0, errors.New("error finding jobs in the database: " + err.Error())
	}
	if len(dbcs) == 0 {
		return 0, errors.New("found zero jobs in the database")
	}

	numRescored := 0
	for _, dbc := range dbcs {
		scoring.ScoreDBComment(dbc)
		err = db.UpsertJob(dbc)
		if err != nil {
			return numRescored, errors.New("failed to upsert the db comment: " + err.Error())
		}
		numRescored++
	}

	return numRescored, nil
}
