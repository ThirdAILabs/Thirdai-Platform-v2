package deployment

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"thirdai_platform/search/ndb"

	"github.com/google/uuid"
)

type LLMCache struct {
	Ndb       ndb.NeuralDB
	Threshold float64
}

func NewLLMCache(modelBazaarDir, modelId string) (*LLMCache, error) {
	cachePath := filepath.Join(modelBazaarDir, "models", modelId, "llm_cache", "llm_cache.ndb")
	ndb, err := ndb.New(cachePath)
	if err != nil {
		return nil, fmt.Errorf("unable to construct LLM Cache: %v", err)
	}

	threshold := 0.95

	return &LLMCache{Ndb: ndb, Threshold: threshold}, nil
}

func (c *LLMCache) Close() {
	c.Ndb.Free()
}

func (c *LLMCache) Suggestions(query string) ([]string, error) {
	slog.Info("fetching cache suggestions", "query", query)

	chunks, err := c.Ndb.Query(query, 5, nil)
	if err != nil {
		return []string{}, fmt.Errorf("ndb query error: %v", err)
	}

	if len(chunks) == 0 {
		slog.Info("cache is empty")
		return []string{}, nil
	}

	seen := make(map[string]struct{})
	suggestions := []string{}

	for _, chunk := range chunks {
		str := chunk.Text
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}
			suggestions = append(suggestions, str)
		}
	}

	return suggestions, nil
}

func tokenSimilarity(queryTokens map[string]struct{}, cachedQuery string) float64 {
	if len(queryTokens) == 0 {
		return 0.0
	}

	cachedTokens := strings.Fields(cachedQuery)
	overlap := 0
	for _, token := range cachedTokens {
		if _, exists := queryTokens[token]; exists {
			overlap++
		}
	}

	return float64(overlap) / float64(len(queryTokens))
}

func (c *LLMCache) Query(query string) (string, error) {
	slog.Info("executing cache request", "query", query)

	chunks, err := c.Ndb.Query(query, 5, nil)
	if err != nil {
		return "", fmt.Errorf("ndb query error: %v", err)
	}

	if len(chunks) == 0 {
		slog.Info("cache is empty")
		return "", fmt.Errorf("cache is empty: %v", err)
	}

	queryTokens := make(map[string]struct{})
	for _, token := range strings.Fields(query) {
		queryTokens[token] = struct{}{}
	}

	sort.Slice(chunks, func(i, j int) bool {
		return tokenSimilarity(queryTokens, chunks[i].Text) > tokenSimilarity(queryTokens, chunks[j].Text)
	})

	topChunk := chunks[0]
	topSimilarity := tokenSimilarity(queryTokens, topChunk.Text)

	if topSimilarity < c.Threshold {
		slog.Info("top chunk similarity below threshold", "similarity", topSimilarity, "threshold", c.Threshold)
		return "", nil
	}

	llmResUncasted, ok := topChunk.Metadata["llm_res"]
	if !ok {
		return "", fmt.Errorf("llm_res metadata value not found: %v", err)
	}

	llmRes, ok := llmResUncasted.(string)
	if !ok {
		return "", fmt.Errorf("llm_res metadata value not of string type: %v", err)
	}

	return llmRes, nil
}

func (c *LLMCache) Insert(query, llmRes string) error {
	slog.Info("inserting to cache", "query", query, "llm_res", llmRes)

	err := c.Ndb.Insert(
		"cache_query", uuid.New().String(),
		[]string{query},
		[]map[string]interface{}{{"llm_res": llmRes}},
		nil)

	if err != nil {
		return fmt.Errorf("failed insertion to cache")
	}

	return nil
}
