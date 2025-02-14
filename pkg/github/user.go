package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type UserResponse struct {
	Data struct {
		Search struct {
			IssueCount int `json:"issueCount"`
			PageInfo   struct {
				EndCursor   string `json:"endCursor"`
				StartCursor string `json:"startCursor"`
			} `json:"pageInfo"`
			Nodes []PullRequestWithReviewsAndChecks `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

type PullRequestWithReviewsAndChecks struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Title             string            `json:"title"`
	Url               string            `json:"url"`
	StatusCheckRollup StatusCheckRollup `json:"statusCheckRollup"`
	Reviews           Reviews           `json:"reviews"`
}

type StatusCheckRollup struct {
	State    string `json:"state"`
	Contexts struct {
		Nodes []map[string]interface{} `json:"nodes"`
	} `json:"contexts"`
}

type CheckRun struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completedAt"`
	Conclusion  string     `json:"conclusion"`
}

type StatusCheck struct {
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	Creator   struct {
		Login string `json:"login"`
	} `json:"creator"`
	Context string `json:"context"`
}

type Reviews struct {
	TotalCount int      `json:"totalCount"`
	Nodes      []Review `json:"nodes"`
}

type Review struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	UpdatedAt time.Time `json:"updatedAt"`
	State     string    `json:"state"`
}

// Output of the function
type Output struct {
	PullRequest PullRequestWithReviewsAndChecks
	Checks      []CheckRun
	Reviews     []Review
}

func (github *GitHub) GetNewReviewsOrNewChecks(author string, since time.Time) (*[]Output, error) {
	logger := slog.Default()
	searchQuery := fmt.Sprintf("is:open is:pr author:%s", author)
	logger.Debug("getNewReviewsOrNewChecks", "searchQuery", searchQuery)
	query := QueryRequestBody{
		Query: `query {
  search(query: "` + searchQuery + `", type: ISSUE, first: 10) {
    issueCount
    nodes {
      ... on PullRequest {
        id
        title
        url
        author {
          login
        }
        statusCheckRollup {
          state
          contexts(first: 10) {
            checkRunCount
            statusContextCount
						nodes {
							... on CheckRun {
								id
								name
								text
								title
								status
								completedAt
								conclusion
							}
							... on StatusContext {
								id
								state
								createdAt
								creator {
									login
								}
								context
							}
						}
          }
        }
        reviews(first: 10) {
          totalCount
          nodes {
						author {
							login
						}
						lastEditedAt
						updatedAt
						submittedAt
						state
          }
        }
      }
    }
  }
}`,
	}
	jsonStr, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonStr))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", github.accessToken))
	if err != nil {
		return nil, err
	}
	res, err := github.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Status code is %s instead of 200", res.Status)
	}
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var body UserResponse
	err = json.Unmarshal(responseBody, &body)
	if err != nil {
		return nil, err
	}

	var outputs []Output
	for _, node := range body.Data.Search.Nodes {
		output := Output{
			PullRequest: PullRequestWithReviewsAndChecks{
				Author: node.Author,
				Title:  node.Title,
				Url:    node.Url,
			},
		}
		var checks []CheckRun
		for _, check := range node.StatusCheckRollup.Contexts.Nodes {
			if strings.HasPrefix(check["id"].(string), "CR_") { // It's a status check
				var completedAt *time.Time
				if check["completedAt"] == nil {
					completedAt = nil
				} else {
					_completedAt, err := time.Parse(time.RFC3339, check["completedAt"].(string))
					if err != nil {
						return nil, err
					}
					completedAt = &_completedAt
				}
				var conclusion string = ""
				if check["conclusion"] != nil {
					conclusion = check["conclusion"].(string)
				}
				checks = append(checks, CheckRun{
					Status:      check["status"].(string),
					CompletedAt: completedAt,
					Name:        check["name"].(string),
					Conclusion:  conclusion,
				})
			} else if strings.HasPrefix(check["id"].(string), "SC_") { // It's a check run
				// SC when in progress it has `PENDING` state, then `COMPLETED` when finished
				completedAt, err := time.Parse(time.RFC3339, check["createdAt"].(string))
				if err != nil {
					return nil, err
				}

				// Exit if the there is a check that is still pending
				if check["state"].(string) == "PENDING" {
					break
				}

				checks = append(checks, CheckRun{
					Status:      "COMPLETED",
					CompletedAt: &completedAt,
					Name:        check["context"].(string),
					Conclusion:  check["state"].(string),
				})
			} else {
				logger.Debug("getNewReviewsOrNewChecks", "check", check)
			}
		}

		var reviews []Review
		for _, review := range node.Reviews.Nodes {
			reviews = append(reviews, Review{
				Author:    review.Author,
				UpdatedAt: review.UpdatedAt,
				State:     review.State,
			})
		}

		// Checking if the checks or reviews are new
		for _, check := range checks {
			if check.CompletedAt != nil && check.CompletedAt.After(since) {
				output.Checks = append(output.Checks, check)
			}
		}
		for _, review := range reviews {
			if review.UpdatedAt.After(since) {
				output.Reviews = append(output.Reviews, review)
			}
		}

		// If we have checks or reviews, we add to the final outputs
		if len(output.Checks) > 0 || len(output.Reviews) > 0 {
			outputs = append(outputs, output)
		}
	}

	return &outputs, nil
}
