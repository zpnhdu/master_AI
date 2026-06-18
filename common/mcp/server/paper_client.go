package mcp

import "context"

type Paper struct {
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	Year     int      `json:"year"`
	Abstract string   `json:"abstract"`
	URL      string   `json:"url"`
	Source   string   `json:"source"`
}

type PaperSearchClient interface {
	SearchPapers(ctx context.Context, query string, limit int) ([]Paper, error)
}
