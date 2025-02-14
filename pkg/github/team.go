package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	Data struct {
		Search struct {
			IssueCount int `json:"issueCount"`
			PageInfo   struct {
				EndCursor   string `json:"endCursor"`
				StartCursor string `json:"startCursor"`
			} `json:"pageInfo"`
			Edges []struct {
				Node PullRequest `json:"node"`
			} `json:"edges"`
		} `json:"search"`
	} `json:"data"`
}

type PullRequest struct {
	Author struct {
		Login string `json:"login"`
	}
	Title     string    `json:"title"`
	Url       string    `json:"url"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (github *GitHub) searchPullRequests(searchQuery string) (*[]PullRequest, error) {
	logger := slog.Default()
	logger.Debug("Search query", "value", searchQuery)
	query := QueryRequestBody{
		Query: `query { 
			search(query: "` + searchQuery + `", type: ISSUE, first: 100) {
				issueCount
				pageInfo {
					endCursor
					startCursor
				}
				edges {
					node {
						... on PullRequest {
							author{
								login
							}
							title
							updatedAt
							url
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
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", github.accessToken))
	res, err := github.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(fmt.Sprintf("Status code is %s instead of 200", res.Status))
	}

	var body Response
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, err
	}

	var pullRequests []PullRequest
	for _, edge := range body.Data.Search.Edges {
		pullRequests = append(pullRequests, edge.Node)
	}
	return &pullRequests, nil
}

func (github *GitHub) GetPullRequests(organization string, team string, me string, authors []string, updatedAt time.Time) (*[]PullRequest, error) {
	var strAuthors []string
	for _, author := range authors {
		strAuthors = append(strAuthors, fmt.Sprintf("author:%s", author))
	}

	searchQuery := fmt.Sprintf(`is:pr -is:draft is:open review-requested:%s org:%s archived:false %s updated:>=%s`, me, organization, strings.Join(strAuthors, " "), updatedAt.UTC().Format(time.RFC3339))

	pullRequests, err := github.searchPullRequests(searchQuery)
	if err != nil {
		return nil, err
	}

	// Add PRs from private users
	allUsersQuery := fmt.Sprintf(`is:pr -is:draft is:open team-review-requested:%s org:%s archived:false updated:>=%s`, team, organization, updatedAt.UTC().Format(time.RFC3339))
	allUsersPRs, err := github.searchPullRequests(allUsersQuery)
	if err != nil {
		return nil, err
	}
	for _, pr := range *allUsersPRs {
		if Contains(authors, pr.Author.Login) {
			// Check if it already exists in the list
			exists := false
			for _, pullRequest := range *pullRequests {
				if pullRequest.Url == pr.Url {
					exists = true
					break
				}
			}
			if exists {
				continue
			}
			*pullRequests = append(*pullRequests, pr)
		}
	}

	return pullRequests, nil
}
