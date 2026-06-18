package memory

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"GopherAI/common/redis"
	"GopherAI/config"

	redisCli "github.com/redis/go-redis/v9"
)

const (
	memoryVectorField = "vector"
	defaultChunkSize  = 700
	defaultOverlap    = 80
)

// RedisVectorStore is the semantic recall layer. It stores chunk hashes:
// memory:{user_id}:{category}:{chunk_id}
// with content/category/user_id/created_at/metadata/vector fields.
type RedisVectorStore struct {
	client    *redisCli.Client
	embedder  MemoryEmbedder
	dimension int
}

func NewRedisVectorStore() (*RedisVectorStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if redis.Rdb == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	if err := redis.Rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis unavailable: %w", err)
	}

	embedder, err := NewArkMemoryEmbedder(ctx)
	if err != nil {
		return nil, fmt.Errorf("embedding unavailable: %w", err)
	}

	return &RedisVectorStore{
		client:    redis.Rdb,
		embedder:  embedder,
		dimension: config.GetConfig().RagModelConfig.RagDimension,
	}, nil
}

func (s *RedisVectorStore) WriteMemory(ctx context.Context, userID, category, content string) error {
	userID = cleanPart(userID)
	category = normalizeCategory(category)
	content = strings.TrimSpace(content)
	if userID == "" || content == "" {
		return fmt.Errorf("userID and content are required")
	}
	if !allowedCategories[category] {
		return fmt.Errorf("unsupported memory category: %s", category)
	}

	chunks := chunkMemory(content, defaultChunkSize, defaultOverlap)
	if len(chunks) == 0 {
		return nil
	}
	if err := s.ensureIndex(ctx, userID); err != nil {
		return err
	}

	vectors, err := s.embedder.Embed(ctx, chunks)
	if err != nil {
		return err
	}
	now := time.Now().Format(time.RFC3339)
	for i, chunk := range chunks {
		if i >= len(vectors) || len(vectors[i]) == 0 {
			continue
		}
		metadata := map[string]string{
			"chunk_index": strconv.Itoa(i),
			"source":      "markdown_memory",
		}
		metadataJSON, _ := json.Marshal(metadata)
		key := fmt.Sprintf("memory:%s:%s:%d-%d", userID, category, time.Now().UnixNano(), i)
		vectorBytes, err := float64ToFloat32Bytes(vectors[i])
		if err != nil {
			return err
		}
		if err := s.client.HSet(ctx, key, map[string]any{
			"content":    chunk,
			"category":   category,
			"user_id":    userID,
			"created_at": now,
			"metadata":   string(metadataJSON),
			"vector":     vectorBytes,
		}).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisVectorStore) SearchMemory(ctx context.Context, userID, query string, limit int) ([]SearchResult, error) {
	userID = cleanPart(userID)
	query = strings.TrimSpace(query)
	if userID == "" || query == "" {
		return nil, fmt.Errorf("userID and query are required")
	}
	if limit <= 0 {
		limit = 5
	}
	if err := s.ensureIndex(ctx, userID); err != nil {
		return nil, err
	}

	vectors, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return nil, fmt.Errorf("empty query vector")
	}
	vectorBytes, err := float64ToFloat32Bytes(vectors[0])
	if err != nil {
		return nil, err
	}

	searchQuery := fmt.Sprintf("*=>[KNN %d @%s $vec AS score]", limit, memoryVectorField)
	raw, err := s.client.Do(ctx,
		"FT.SEARCH", memoryIndexName(userID), searchQuery,
		"PARAMS", "2", "vec", vectorBytes,
		"RETURN", "5", "content", "category", "user_id", "metadata", "score",
		"SORTBY", "score", "ASC",
		"DIALECT", "2",
	).Result()
	if err != nil {
		return nil, err
	}
	return parseVectorSearchResults(raw), nil
}

func (s *RedisVectorStore) ensureIndex(ctx context.Context, userID string) error {
	indexName := memoryIndexName(userID)
	if err := s.client.Do(ctx, "FT.INFO", indexName).Err(); err == nil {
		return nil
	} else if !isMissingIndexErr(err) {
		return err
	}

	prefix := fmt.Sprintf("memory:%s:", cleanPart(userID))
	return s.client.Do(ctx,
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",
		"content", "TEXT",
		"category", "TAG",
		"user_id", "TAG",
		"created_at", "TEXT",
		"metadata", "TEXT",
		memoryVectorField, "VECTOR", "FLAT", "6",
		"TYPE", "FLOAT32",
		"DIM", s.dimension,
		"DISTANCE_METRIC", "COSINE",
	).Err()
}

func memoryIndexName(userID string) string {
	return "idx:memory:" + cleanPart(userID)
}

func isMissingIndexErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown index") || strings.Contains(msg, "no such index")
}

func chunkMemory(content string, size, overlap int) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	runes := []rune(content)
	if len(runes) <= size {
		return []string{content}
	}
	var chunks []string
	for start := 0; start < len(runes); {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

func float64ToFloat32Bytes(vector []float64) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(vector)*4))
	for _, v := range vector {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			v = 0
		}
		if err := binary.Write(buf, binary.LittleEndian, float32(v)); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func parseVectorSearchResults(raw any) []SearchResult {
	items, ok := raw.([]any)
	if !ok || len(items) < 2 {
		return nil
	}
	results := make([]SearchResult, 0, (len(items)-1)/2)
	for i := 2; i < len(items); i += 2 {
		fields, ok := items[i].([]any)
		if !ok {
			continue
		}
		data := fieldsToMap(fields)
		distance := parseFloat(data["score"])
		score := int((1 - distance) * 1000)
		if score < 0 {
			score = 0
		}
		results = append(results, SearchResult{
			Category: data["category"],
			Content:  data["content"],
			Score:    score,
			Source:   "redis_vector",
			Metadata: map[string]string{
				"distance": fmt.Sprintf("%.6f", distance),
				"user_id":  data["user_id"],
				"raw":      data["metadata"],
			},
		})
	}
	return results
}

func fieldsToMap(fields []any) map[string]string {
	out := make(map[string]string)
	for i := 0; i+1 < len(fields); i += 2 {
		key := fmt.Sprint(fields[i])
		value := fields[i+1]
		switch v := value.(type) {
		case []byte:
			out[key] = string(v)
		default:
			out[key] = fmt.Sprint(v)
		}
	}
	return out
}

func parseFloat(text string) float64 {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 1
	}
	return value
}
