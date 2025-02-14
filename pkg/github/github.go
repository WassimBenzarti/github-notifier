package github

import (
	"net/http"
)

type GitHub struct {
	accessToken string
	client      *http.Client
}

func NewGitHub(accessToken string) *GitHub {
	return &GitHub{
		accessToken: accessToken,
		client:      &http.Client{},
	}
}

type QueryRequestBody struct {
	Query string `json:"query"`
}
