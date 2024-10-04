package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"

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

func askUserForDefaults() {
	viper.Set("me", StringPrompt("What is your GitHub username?"))
	viper.Set("org", StringPrompt("What is your organization?"))
	viper.Set("team", StringPrompt("What is your GitHub team (<org>/<team>)?"))
	viper.Set("team-members", ListPrompt("Type the names of your teammates or people you want to follow (Hit <Enter> after every name, Hit <Enter> twice when the list is done)\n> "))
}

func initConfig() {
	// Get the home directory
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Default config directory path
	configDir := filepath.Join(home, ".config", "github-notifier")
	configFilePath := filepath.Join(configDir, "config.json")

	viper.AddConfigPath(configDir)
	viper.SetConfigType("json")
	viper.SetConfigName("config")

	// Check if the config is overridden
	if cfgFile != "" {
		workingdir, err := os.Getwd()
		cobra.CheckErr(err)

		configFilePath = path.Join(workingdir, cfgFile)
		// Use config file from the flag.
		viper.SetConfigFile(configFilePath)
		slog.Debug("Using the specified config file under", "path", configFilePath)
	}

	os.MkdirAll(configDir, os.ModePerm)
	askUserForDefaults()
	slog.Info("Config file successfully created", "path", configFilePath)

	err = viper.SafeWriteConfig()
	cobra.CheckErr(err)
}

var initCommand = &cobra.Command{
	Use:   "init",
	Short: "(Re)initialize the config file",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
}

func init() {
	configCommand.AddCommand(initCommand)
	configCommand.AddCommand(editCommand)
	rootCmd.AddCommand(configCommand)
}
