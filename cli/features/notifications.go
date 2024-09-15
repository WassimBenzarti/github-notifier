package features

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/wassimbenzarti/github-notifier/github"
	"github.com/wassimbenzarti/github-notifier/terminal"
)

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

func RunNotifications() {
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

		pullRequests, err := githubClient.GetPullRequests("egym",
			"egym/sre",
			"wassimbenzarti",
			[]string{"wassimbenzarti", "eldad", "leonnicolas", "eldad", "kolsware", "Akaame", "gjabell", "martykuentzel", "viktorkholod", "jhandguy", "kostya2011"},
			createdAt,
		)
		if err != nil {
			panic(err)
		}

		var messages []string
		if len(*pullRequests) > 0 {
			messages = append(messages, fmt.Sprintf("You have %d new PR(s) to review", len(*pullRequests)))
			for _, pullRequest := range *pullRequests {
				terminal.ColorfulPrintf(terminal.Blue, "REVIEW: %s\t%s\t%s\n", pullRequest.Author.Login, pullRequest.Title, pullRequest.Url)
			}
		} else {
			slog.Debug("No new notifications")
		}

		myPullrequests, err := githubClient.GetNewReviewsOrNewChecks("@me", createdAt)
		if err != nil {
			panic(err)
		}

		if len(*myPullrequests) > 0 {
			messages = append(messages, fmt.Sprintf("You have %d new review(s) or check(s)", len(*myPullrequests)))
			for _, pr := range *myPullrequests {

				terminal.ColorfulPrintf(terminal.Green, "PR title: %s\t%s\n", pr.PullRequest.Title, pr.PullRequest.Url)
				for _, review := range pr.Reviews {
					terminal.ColorfulPrintf(terminal.Green, "\tReview by: %s\n", review.Author.Login)
				}
				for _, check := range pr.Checks {
					terminal.ColorfulPrintf(terminal.Green, "\tCompleted check: %s\n", check.Name)
				}
			}
		}
		if len(messages) > 0 {
			beeep.Alert("GH Notifier", strings.Join(messages, "\n"), "assets/notification.png")
		}
	}

}
