package cmd

import (
	"context"
	"fmt"
	"github.com/mwinters0/hnjobs/app"
	"github.com/mwinters0/hnjobs/config"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch without loading the UI",
	Long:  `Fetch and score the latest job story and comments without loading the UI.`,
	Run:   fetch,
}

var flagForce bool
var flagQuiet bool
var flagStoryID int
var flagExit bool

func init() {
	rootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().BoolVarP(
		&flagForce,
		"force", "f",
		false,
		"Fetch and score all comments, ignoring cache TTL",
	)
	fetchCmd.Flags().BoolVarP(
		&flagQuiet,
		"quiet", "q",
		false,
		"Don't print status info. Might still print errors.",
	)
	fetchCmd.Flags().IntVarP(
		&flagStoryID,
		"storyid", "s",
		0,
		"Fetch a specific story ID instead of latest",
	)
	fetchCmd.Flags().BoolVarP(
		&flagExit,
		"exit", "x",
		false,
		"Exit code is 0 only if new jobs were fetched. Empty success is code 42.",
	)
}

func fetch(cmd *cobra.Command, args []string) {
	status := make(chan app.FetchStatusUpdate)
	fo := app.FetchOptions{
		Context:     context.Background(),
		Status:      status,
		ModeForce:   flagForce,
		StoryID:     flagStoryID,
		TTLSec:      config.GetConfig().Cache.TTLSecs,
		MustContain: app.WhoIsHiringString,
	}
	go app.FetchAsync(fo)
	for {
		select {
		case fsu, ok := <-status:
			if !ok {
				//EOF - it's a bug if we see this
				log.Fatal("BUG: cmd/fetch: status channel closed before UpdateTypeDone!")
			}
			if !flagQuiet {
				fmt.Println(fsu.Message)
			}
			switch fsu.UpdateType {
			case app.UpdateTypeFatal:
				if !flagQuiet {
					fmt.Println("Fetch experienced fatal errors.")
				}
				os.Exit(1)
			case app.UpdateTypeGeneric,
				app.UpdateTypeNewStory,
				app.UpdateTypeNonFatalErr,
				app.UpdateTypeBadComment,
				app.UpdateTypeJobFetched:
			case app.UpdateTypeDone:
				// This is where we intend to exit
				if flagExit && fsu.Value == 0 {
					// no new jobs fetched
					os.Exit(42)
				}
				return
			default:
				log.Fatal(fmt.Sprintf("BUG: cmd/fetch: unhandled UpdateType %d", fsu.UpdateType))
			}
		}
	}
}
