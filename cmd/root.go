package cmd

import (
	"github.com/spf13/cobra"
	"hnjobs/app"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "hnjobs",
	Short: "Making tomorrow's world a better place today",
	Long: `hnjobs: making tomorrow's world a better place today

Just run the app without any commands / flags unless you think you're special.  Press F1 in the TUI for help.

`,
	Run: browse,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//rootCmd.Flags().BoolVarP(&app.BrowseOptions.MouseEnabled, "mouse", "m", true, "Set TTY mouse enabled (default --mouse=true)")
}

func browse(cmd *cobra.Command, args []string) {
	app.Browse()
	return
}
