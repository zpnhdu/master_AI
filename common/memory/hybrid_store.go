package memory

import (
	"context"
	"log"
)

type VectorMemoryStore interface {
	WriteMemory(ctx context.Context, userID, category, content string) error
	SearchMemory(ctx context.Context, userID, query string, limit int) ([]SearchResult, error)
}

// HybridStore keeps Memory durable and inspectable in Markdown, then attempts
// a best-effort Redis Vector write for semantic recall. Vector failures never
// block the main app path.
type HybridStore struct {
	markdown *MarkdownStore
	vector   VectorMemoryStore
}

func NewHybridStore(markdown *MarkdownStore, vector VectorMemoryStore) *HybridStore {
	return &HybridStore{markdown: markdown, vector: vector}
}

func (s *HybridStore) WriteMemory(userID, category, content string) error {
	if err := s.markdown.WriteMemory(userID, category, content); err != nil {
		return err
	}
	if s.vector == nil {
		return nil
	}
	if err := s.vector.WriteMemory(context.Background(), userID, category, content); err != nil {
		log.Printf("[memory] vector write skipped: %v", err)
	}
	return nil
}

func (s *HybridStore) ReadMemory(userID, category string) (string, error) {
	return s.markdown.ReadMemory(userID, category)
}

func (s *HybridStore) SearchMemory(userID, query string) ([]SearchResult, error) {
	if s.vector != nil {
		results, err := s.vector.SearchMemory(context.Background(), userID, query, 5)
		if err == nil && len(results) > 0 {
			return results, nil
		}
		if err != nil {
			log.Printf("[memory] vector search fallback to markdown: %v", err)
		}
	}
	return s.markdown.SearchMemory(userID, query)
}
