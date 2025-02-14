package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/wassimbenzarti/github-notifier/pkg/features"
)

var runCommand = &cobra.Command{
	Use:   "run",
	Short: "Start the github-notifier",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if the configuration is valid
		err := viper.ReadInConfig()
		cobra.CheckErr(err)

		features.UseNotifySend = UseNotifySend

		features.RunNotifications(
			viper.GetString("org"),
			viper.GetString("team"),
			viper.GetString("me"),
			viper.GetStringSlice("team-members"),
		)
	},
}

func init() {
	rootCmd.AddCommand(runCommand)
}
