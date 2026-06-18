package memory

import (
	"context"
	"fmt"
	"os"

	"GopherAI/config"

	embeddingArk "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
)

type MemoryEmbedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// ArkMemoryEmbedder reuses the same Ark embedding configuration used by RAG.
// It intentionally reads config.RagModelConfig and OPENAI_API_KEY instead of
// hardcoding provider credentials in the memory package.
type ArkMemoryEmbedder struct {
	embedder embedding.Embedder
}

func NewArkMemoryEmbedder(ctx context.Context) (*ArkMemoryEmbedder, error) {
	cfg := config.GetConfig()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is empty")
	}

	embedder, err := embeddingArk.NewEmbedder(ctx, &embeddingArk.EmbeddingConfig{
		BaseURL: cfg.RagModelConfig.RagBaseUrl,
		APIKey:  apiKey,
		Model:   cfg.RagModelConfig.RagEmbeddingModel,
	})
	if err != nil {
		return nil, err
	}
	return &ArkMemoryEmbedder{embedder: embedder}, nil
}

func (e *ArkMemoryEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	return e.embedder.EmbedStrings(ctx, texts)
}
