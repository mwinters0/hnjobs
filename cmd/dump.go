package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mwinters0/hnjobs/db"
	"github.com/mwinters0/hnjobs/hn"
	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump the current month's data to stdout as json",
	Long:  "Dump the current month's data to stdout as json",
	Run:   dump,
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}

type dumpData struct {
	Story *hn.Story
	Jobs  []*db.Job
}

func dump(cmd *cobra.Command, args []string) {
	latest, err := db.GetLatestStory()
	if errors.Is(err, db.ErrNoResults) {
		panic("No stories found")
		return
	}
	if err != nil {
		panic(fmt.Errorf("error finding latest job story from DB: %v", err))
	}
	jobs, err := db.GetAllJobsByStoryId(latest.Id, db.OrderScoreDesc)
	if err != nil {
		panic(fmt.Sprintf("error getting jobs from DB: %v", err))
	}
	if len(jobs) == 0 {
		panic(fmt.Sprintf("No jobs in DB for latest story ID %d (%s)", latest.Id, latest.Title))
	}

	d := &dumpData{
		Story: latest,
		Jobs:  jobs,
	}
	j, err := json.Marshal(d)
	if err != nil {
		panic("Error marshaling JSON")
	}
	fmt.Println(string(j))
}
