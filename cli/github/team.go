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
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (github *GitHub) searchPullRequests(searchQuery string) (*[]PullRequest, error) {
	logger := slog.Default()
	logger.Debug("Search query", "value", searchQuery)
	var query = QueryRequestBody{
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", github.accessToken))
	if err != nil {
		return nil, err
	}
	res, err := github.client.Do(req)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf(fmt.Sprintf("Status code is %s instead of 200", res.Status))
	}
	if err != nil {
		return nil, err
	}
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var body Response
	err = json.Unmarshal(responseBody, &body)
	if err != nil {
		return nil, err
	}
	var pullRequests []PullRequest
	for _, edge := range body.Data.Search.Edges {
		pullRequests = append(pullRequests, edge.Node)
	}
	return &pullRequests, nil
}

func (github *GitHub) GetPullRequests(organization string, team string, me string, authors []string, createdAt time.Time) (*[]PullRequest, error) {

	var strAuthors []string
	for _, author := range authors {
		strAuthors = append(strAuthors, fmt.Sprintf("author:%s", author))
	}

	searchQuery := fmt.Sprintf(`is:pr is:open review-requested:%s org:%s archived:false %s created:>=%s`, me, organization, strings.Join(strAuthors, " "), createdAt.UTC().Format(time.RFC3339))

	pullRequests, err := github.searchPullRequests(searchQuery)
	if err != nil {
		return nil, err
	}

	// Add PRs from private users
	allUsersQuery := fmt.Sprintf(`is:pr is:open team-review-requested:%s org:%s archived:false created:>=%s`, team, organization, createdAt.UTC().Format(time.RFC3339))
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
