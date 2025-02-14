package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Show the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v0.0.2")
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
}
