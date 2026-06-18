package mcp

import "context"

type Job struct {
	Company      string   `json:"company"`
	Title        string   `json:"title"`
	City         string   `json:"city"`
	Requirements []string `json:"requirements"`
	URL          string   `json:"url"`
}

type JobSearchClient interface {
	SearchJobs(ctx context.Context, keyword, city string, limit int) ([]Job, error)
}
