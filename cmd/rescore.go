package cmd

import (
	"errors"
	"fmt"
	"github.com/mwinters0/hnjobs/app"
	"github.com/mwinters0/hnjobs/db"
	"github.com/spf13/cobra"
	"log"
)

var rescoreCmd = &cobra.Command{
	Use:   "rescore",
	Short: "Re-score all jobs in the database without fetching (useful if your rules changed)",
	Long:  `Re-score all jobs in the database without fetching (useful if your rules changed)`,
	Run:   rescore,
}

func init() {
	rootCmd.AddCommand(rescoreCmd)
}

func rescore(cmd *cobra.Command, args []string) {
	stories, err := db.GetAllStories()
	if err != nil && !errors.Is(err, db.ErrNoResults) {
		log.Fatal("Error getting stories from DB: " + err.Error())
	}
	if errors.Is(err, db.ErrNoResults) {
		fmt.Println("No stories found")
		return
	}

	numRescored := 0
	if len(stories) == 1 {
		fmt.Printf("Found %d story, rescoring...\n", len(stories))
	} else {
		fmt.Printf("Found %d stories, rescoring...\n", len(stories))
	}
	for _, story := range stories {
		num, err := app.ReScore(story.Id)
		if err != nil {
			log.Fatal("ERROR: " + err.Error())
		}
		numRescored += num
	}
	fmt.Printf("Rescored %d jobs\n", numRescored)
}
