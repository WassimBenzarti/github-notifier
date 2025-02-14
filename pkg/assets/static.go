package assets

import (
	"embed"
	"io"
	"log/slog"
	"os"
)

//go:embed notification.png
var assets embed.FS

var NotificationIconFilePath string

func init() {
	file, err := os.CreateTemp(os.TempDir(), "github-notifier-*.png")
	if err != nil {
		slog.Error("failed to create temp file", "err", err.Error())
		return
	}
	defer file.Close()

	sourceFile, err := assets.Open("notification.png")
	if err != nil {
		slog.Error("failed to open asset file", "err", err.Error())
		return
	}
	defer sourceFile.Close()

	_, err = io.Copy(file, sourceFile)
	if err != nil {
		slog.Error("failed to copy asset file", "err", err.Error())
		return
	}

	NotificationIconFilePath = file.Name()
}
