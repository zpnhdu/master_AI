package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type EmbeddingStore interface {
	Search(userID, query string, limit int) ([]SearchResult, error)
}

type SearchResult struct {
	Category string            `json:"category"`
	Content  string            `json:"content"`
	Score    int               `json:"score"`
	Source   string            `json:"source,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Store keeps the old public API while delegating to HybridStore.
// Callers still use WriteMemory / ReadMemory / SearchMemory exactly as before.
type Store struct {
	hybrid *HybridStore
}

// MarkdownStore is the readable, editable and auditable memory layer.
// It stores one markdown file per user/category under data/memory/{user_id}.
type MarkdownStore struct {
	baseDir string
}

var allowedCategories = map[string]bool{
	"profile":     true,
	"research":    true,
	"career":      true,
	"preferences": true,
}

func NewStore(baseDir string) *Store {
	markdown := NewMarkdownStore(baseDir)
	vectorStore, _ := NewRedisVectorStore()
	return &Store{
		hybrid: NewHybridStore(markdown, vectorStore),
	}
}

func (s *Store) WriteMemory(userID, category, content string) error {
	return s.hybrid.WriteMemory(userID, category, content)
}

func (s *Store) ReadMemory(userID, category string) (string, error) {
	return s.hybrid.ReadMemory(userID, category)
}

func (s *Store) SearchMemory(userID, query string) ([]SearchResult, error) {
	return s.hybrid.SearchMemory(userID, query)
}

func NewMarkdownStore(baseDir string) *MarkdownStore {
	if baseDir == "" {
		baseDir = "./data/memory"
	}
	return &MarkdownStore{baseDir: baseDir}
}

func (s *MarkdownStore) WriteMemory(userID, category, content string) error {
	userID = cleanPart(userID)
	category = normalizeCategory(category)
	content = strings.TrimSpace(content)
	if userID == "" {
		return fmt.Errorf("userID is required")
	}
	if content == "" {
		return fmt.Errorf("content is required")
	}
	if !allowedCategories[category] {
		return fmt.Errorf("unsupported memory category: %s", category)
	}

	dir := filepath.Join(s.baseDir, userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, category+".md")
	entry := fmt.Sprintf("\n\n## %s\n%s\n", time.Now().Format(time.RFC3339), content)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(entry)
	return err
}

func (s *MarkdownStore) ReadMemory(userID, category string) (string, error) {
	userID = cleanPart(userID)
	category = normalizeCategory(category)
	if userID == "" {
		return "", fmt.Errorf("userID is required")
	}
	if !allowedCategories[category] {
		return "", fmt.Errorf("unsupported memory category: %s", category)
	}

	content, err := os.ReadFile(filepath.Join(s.baseDir, userID, category+".md"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *MarkdownStore) SearchMemory(userID, query string) ([]SearchResult, error) {
	userID = cleanPart(userID)
	query = strings.TrimSpace(query)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	terms := splitTerms(query)
	var results []SearchResult
	for category := range allowedCategories {
		content, err := s.ReadMemory(userID, category)
		if err != nil {
			return nil, err
		}
		score := scoreText(content, terms)
		if score > 0 {
			results = append(results, SearchResult{
				Category: category,
				Content:  trimForResult(content, 600),
				Score:    score,
				Source:   "markdown",
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results, nil
}

func normalizeCategory(category string) string {
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return "preferences"
	}
	return category
}

func cleanPart(part string) string {
	part = strings.TrimSpace(part)
	part = strings.ReplaceAll(part, "\\", "_")
	part = strings.ReplaceAll(part, "/", "_")
	part = strings.ReplaceAll(part, "..", "_")
	return part
}

func splitTerms(text string) []string {
	fields := strings.Fields(strings.ToLower(text))
	if len(fields) == 0 && strings.TrimSpace(text) != "" {
		return []string{strings.ToLower(strings.TrimSpace(text))}
	}
	return fields
}

func scoreText(text string, terms []string) int {
	lower := strings.ToLower(text)
	score := 0
	for _, term := range terms {
		if term == "" {
			continue
		}
		score += strings.Count(lower, term)
	}
	return score
}

func trimForResult(text string, max int) string {
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return text[:max] + "..."
}
