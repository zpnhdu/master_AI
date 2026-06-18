package mcp

import (
	"context"
	"fmt"
	"strings"
)

type MockPaperSearchClient struct{}

func NewMockPaperSearchClient() *MockPaperSearchClient {
	return &MockPaperSearchClient{}
}

func (c *MockPaperSearchClient) SearchPapers(ctx context.Context, query string, limit int) ([]Paper, error) {
	limit = clampLimit(limit)
	topics := []string{"Agentic RAG", "tool-augmented retrieval", "scientific workflow planning", "LLM memory", "research assistant"}
	papers := make([]Paper, 0, limit)
	for i := 0; i < limit; i++ {
		topic := topics[i%len(topics)]
		papers = append(papers, Paper{
			Title:    fmt.Sprintf("%s for %s: A Practical Study", strings.Title(topic), query),
			Authors:  []string{"Chen Li", "Yuki Tanaka", "Maria Smith"},
			Year:     2023 + i%3,
			Abstract: fmt.Sprintf("This mock paper studies %s in the context of %s, focusing on retrieval planning, evidence grounding, and reproducible evaluation.", topic, query),
			URL:      fmt.Sprintf("https://example.org/papers/%s/%d", slug(query), i+1),
			Source:   "mock-arxiv",
		})
	}
	return papers, nil
}

type MockGitHubSearchClient struct{}

func NewMockGitHubSearchClient() *MockGitHubSearchClient {
	return &MockGitHubSearchClient{}
}

func (c *MockGitHubSearchClient) SearchRepos(ctx context.Context, query, language string, limit int) ([]Repo, error) {
	limit = clampLimit(limit)
	if language == "" {
		language = "Go"
	}
	repos := make([]Repo, 0, limit)
	for i := 0; i < limit; i++ {
		repos = append(repos, Repo{
			Repo:        fmt.Sprintf("gradpilot/%s-%s-%d", slug(query), strings.ToLower(language), i+1),
			Description: fmt.Sprintf("%s implementation for %s with README, dataset notes, and reproducible scripts.", language, query),
			Stars:       320 + i*137,
			Language:    language,
			URL:         fmt.Sprintf("https://github.com/gradpilot/%s-%d", slug(query), i+1),
		})
	}
	return repos, nil
}

type MockJobSearchClient struct{}

func NewMockJobSearchClient() *MockJobSearchClient {
	return &MockJobSearchClient{}
}

func (c *MockJobSearchClient) SearchJobs(ctx context.Context, keyword, city string, limit int) ([]Job, error) {
	limit = clampLimit(limit)
	if city == "" {
		city = "remote"
	}
	jobs := make([]Job, 0, limit)
	for i := 0; i < limit; i++ {
		jobs = append(jobs, Job{
			Company:      fmt.Sprintf("AI Lab %d", i+1),
			Title:        fmt.Sprintf("%s Engineer", strings.Title(keyword)),
			City:         city,
			Requirements: []string{"Go/Gin 后端经验", "RAG 或 LLM 应用经验", "Redis/MySQL/RabbitMQ 工程实践", "能清楚讲项目架构与优化"},
			URL:          fmt.Sprintf("https://jobs.example.com/%s/%d", slug(keyword), i+1),
		})
	}
	return jobs, nil
}
