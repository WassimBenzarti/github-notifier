package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCommand = &cobra.Command{
	Use:   "config",
	Short: "Show the config being used",
	Run: func(cmd *cobra.Command, args []string) {
		payload, err := json.MarshalIndent(viper.AllSettings(), "", "  ")
		if err != nil {
			cobra.CheckErr(err)
		}
		fmt.Printf("%s", payload)
	},
}

var editCommand = &cobra.Command{
	Use:   "edit",
	Short: "Edit the config",
	Run: func(cmd *cobra.Command, args []string) {
		editorCommand := exec.Command("vim", viper.ConfigFileUsed())
		editorCommand.Stdin = os.Stdin
		editorCommand.Stdout = os.Stdout
		if err := editorCommand.Run(); err != nil {
			cobra.CheckErr(err)
		}
	},
}

func init() {
	configCommand.AddCommand(editCommand)
	rootCmd.AddCommand(configCommand)
}
