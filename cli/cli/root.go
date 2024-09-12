package cli

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "github-notifier",
	Short: "GitHub notifier",
	Long:  `A CLI for pushing notifications of important GitHub events (Review required, Review received, Checks done, etc.)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/github-notifier/config.json)")
	rootCmd.PersistentFlags().Bool("init", false, "Will re-initialize the config file")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		workingdir, err := os.Getwd()
		if err != nil {
			cobra.CheckErr(err)
		}

		// Use config file from the flag.
		viper.SetConfigFile(path.Join(workingdir, cfgFile))
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}

		viper.AddConfigPath(path.Join(home, ".config", "github-notifier"))
		viper.SetConfigType("json")
		viper.SetConfigName("config.json")
	}

	if err := viper.ReadInConfig(); err != nil {
		_, isConfigFileNotFound := err.(viper.ConfigFileNotFoundError)
		_, isPathError := err.(*fs.PathError)
		if isConfigFileNotFound || isPathError {
			slog.Info("Config file not found, initializing a new config file with", "path", viper.ConfigFileUsed())
			init, err := rootCmd.Flags().GetBool("init")
			if err != nil {
				cobra.CheckErr(err)
			}
			if init {
				askUserForDefaults()
			}
			err = viper.WriteConfig()
			if err != nil {
				cobra.CheckErr(err)
			}
		} else {
			log.Fatalf("Cannot parse the config file, ensure that the file %s is a valid json", viper.ConfigFileUsed())
		}
	}

}

func askUserForDefaults() {
	viper.Set("me", StringPrompt("What is your GitHub username?"))
	viper.Set("org", StringPrompt("What is your organization?"))
	viper.Set("team", StringPrompt("What is your GitHub team (<org>/<team>)?"))
	viper.Set("team-members", ListPrompt("Type the names of your teammates or people you want to follow (Hit <Enter> after every name, Hit <Enter> twice when the list is done)\n> "))
}

func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func ListPrompt(label string) []string {
	var result []string
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s == "\n" {
			break
		} else {
			result = append(result, strings.TrimSpace(s))
		}
	}
	return result
}
