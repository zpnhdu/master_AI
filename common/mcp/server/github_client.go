package mcp

import "context"

type Repo struct {
	Repo        string `json:"repo"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
	Language    string `json:"language"`
	URL         string `json:"url"`
}

type GitHubSearchClient interface {
	SearchRepos(ctx context.Context, query, language string, limit int) ([]Repo, error)
}
