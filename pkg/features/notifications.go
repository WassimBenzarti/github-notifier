package features

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gen2brain/beeep"

	"github.com/wassimbenzarti/github-notifier/pkg/assets"
	"github.com/wassimbenzarti/github-notifier/pkg/github"
	"github.com/wassimbenzarti/github-notifier/pkg/terminal"
)

var UseNotifySend = false

type QueryRequestBody struct {
	Query string `json:"query"`
}

type PullRequest struct {
	Author struct {
		Login string `json:"login"`
	}
	Title string `json:"title"`
	Url   string `json:"url"`
}

func notifySend(summary string, body string, url string) {
	if !UseNotifySend {
		return
	}

	go func() {
		cmd := exec.Command("notify-send", "-i", assets.NotificationIconFilePath, summary, body, "-A", "OPEN-URL=Open URL")
		b, err := cmd.Output()
		if err != nil {
			slog.Debug("failed to call notify-send", "err", err.Error())
			return
		}

		if strings.TrimSpace(string(b)) == "OPEN-URL" {
			cmd := exec.Command("xdg-open", url)
			cmd.Run()
		}
	}()
}

func RunNotifications(organization string, team string, author string, teamMembers []string) {
	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		panic("GitHub token wasn't provided as a GITHUB_TOKEN enviroment variable.")
	}
	githubClient := github.NewGitHub(accessToken)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	firstTime := true
	for ; true; <-ticker.C {

		createdAt := time.Now().Add(-1*time.Minute - 1*time.Second)

		// Check PRs from the last day for the first fetch
		if firstTime {
			createdAt = time.Now().Add(-24 * time.Hour)
			firstTime = false
		}

		pullRequests, err := githubClient.GetPullRequests(
			organization,
			team,
			author,
			teamMembers,
			createdAt,
		)
		if err != nil {
			slog.Warn("Failed to fetch Pull Requests because of an error. Retrying in 1 min...", "Error", err)
			time.Sleep(60 * time.Second)
			continue
		}

		var messages []string
		if len(*pullRequests) > 0 {
			messages = append(messages, fmt.Sprintf("You have %d new PR(s) to review", len(*pullRequests)))
			for _, pullRequest := range *pullRequests {
				terminal.ColorfulPrintf(terminal.LightBlue, "REVIEW: %s\t%s\t%s\n", pullRequest.Author.Login, pullRequest.Title, pullRequest.Url)
				notifySend(fmt.Sprintf("PR @%s", pullRequest.Author.Login), pullRequest.Title, pullRequest.Url)
			}
		} else {
			slog.Debug("No new notifications")
		}

		myPullrequests, err := githubClient.GetNewReviewsOrNewChecks("@me", createdAt)
		if err != nil {
			slog.Warn("Failed to fetch Pull Requests because of an error. Retrying in 1 min...", "Error", err)
			time.Sleep(60 * time.Second)
			continue
		}

		if len(*myPullrequests) > 0 {
			messages = append(messages, fmt.Sprintf("You have %d new review(s) or check(s)", len(*myPullrequests)))
			for _, pr := range *myPullrequests {

				terminal.ColorfulPrintf(terminal.Green, "PR title: %s\t%s\n", pr.PullRequest.Title, pr.PullRequest.Url)
				details := ""
				for _, review := range pr.Reviews {
					details += fmt.Sprintf("Review by @%s\n", review.Author.Login)
					terminal.ColorfulPrintf(terminal.Green, "\tReview by: %s\n", review.Author.Login)
				}
				for _, check := range pr.Checks {
					details += fmt.Sprintf("Completed check %s\n", check.Name)
					terminal.ColorfulPrintf(terminal.Green, "\tCompleted check: %s\n", check.Name)
				}

				notifySend(fmt.Sprintf("PR Activity for %s", pr.PullRequest.Title), details, pr.PullRequest.Url)
			}
		}
		if len(messages) > 0 {
			if !UseNotifySend {
				beeep.Alert("GH Notifier", strings.Join(messages, "\n"), assets.NotificationIconFilePath)
			}
		}
	}
}
