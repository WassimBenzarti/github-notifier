package cli

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	logLevel      int
	UseNotifySend bool
)

var rootCmd = &cobra.Command{
	Use:   "github-notifier",
	Short: "GitHub notifier",
	Long:  `A CLI for pushing notifications of important GitHub events (Review required, Review received, Checks done, etc.)`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.Level(logLevel),
		})))
	},
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
	cobra.OnInitialize(findDefaultConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/github-notifier/config.json)")
	rootCmd.PersistentFlags().IntVar(&logLevel, "log-level", int(slog.LevelInfo), "log level like defined in https://pkg.go.dev/log/slog#Level")
	rootCmd.PersistentFlags().Bool("init", false, "Will re-initialize the config file")
	rootCmd.PersistentFlags().BoolVar(&UseNotifySend, "notify-send", runtime.GOOS == "linux", "Use notify-send (supports actions; Linux-specific)")
}

func findDefaultConfig() {
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
		if err != nil {
			cobra.CheckErr(err)
		}

		configFilePath = path.Join(workingdir, cfgFile)
		// Use config file from the flag.
		viper.SetConfigFile(configFilePath)
		slog.Debug("Using the specified config file under", "path", configFilePath)
		return
	}
	if _, err := os.Stat(configFilePath); err != nil {
		// slog.Error("It seems that the default config file doesn't exist, Try running `github-notifier config init`.", "path", configFilePath)
		return // We want to allow the user to use the init the command
	}
	err = viper.ReadInConfig()
	cobra.CheckErr(err)
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
