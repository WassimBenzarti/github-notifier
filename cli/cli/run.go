package cli

import (
	"github.com/spf13/cobra"
	"github.com/wassimbenzarti/github-notifier/features"
)

var runCommand = &cobra.Command{
	Use:   "run",
	Short: "Start the github-notifier",
	Run: func(cmd *cobra.Command, args []string) {
		features.RunNotifications()
	},
}

func init() {
	rootCmd.AddCommand(runCommand)
}
