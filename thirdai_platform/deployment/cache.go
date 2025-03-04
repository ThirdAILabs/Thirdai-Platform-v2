package deployment

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"thirdai_platform/search/ndb"
)

type LLMCache struct {
	Ndb       ndb.NeuralDB
	Threshold float64
}

const CacheScoreThreshold = 0.95

func NewLLMCache(modelBazaarDir, modelId string) (*LLMCache, error) {
	cachePath := filepath.Join(modelBazaarDir, "models", modelId, "llm_cache", "llm_cache.ndb")
	ndb, err := ndb.New(cachePath)
	if err != nil {
		return nil, fmt.Errorf("unable to construct LLM Cache: %v", err)
	}

	return &LLMCache{Ndb: ndb, Threshold: CacheScoreThreshold}, nil
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

func getTopChunkSimilarity(query string, chunks []ndb.Chunk) (ndb.Chunk, float64) {
	queryTokens := make(map[string]struct{})
	for _, token := range strings.Fields(query) {
		queryTokens[token] = struct{}{}
	}

	sort.Slice(chunks, func(i, j int) bool {
		return tokenSimilarity(queryTokens, chunks[i].Text) > tokenSimilarity(queryTokens, chunks[j].Text)
	})

	topChunk := chunks[0]
	topSimilarity := tokenSimilarity(queryTokens, topChunk.Text)
	return topChunk, topSimilarity
}

func referenceIdsToString(referenceIds []uint64) string {
	referenceIdString := make([]string, len(referenceIds))
	for i, num := range referenceIds {
		referenceIdString[i] = strconv.FormatUint(num, 10)
	}
	return strings.Join(referenceIdString, " ")
}

func referenceIdsFromString(referenceIdString string) ([]uint64, error) {
	parts := strings.Split(referenceIdString, " ")
	referenceIds := make([]uint64, len(parts))

	for i, part := range parts {
		num, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return []uint64{}, fmt.Errorf("could not parse reference ids from metadata %s", referenceIdString)
		}
		referenceIds[i] = num
	}
	return referenceIds, nil
}

func getChunkMetadata(chunk ndb.Chunk) (string, []uint64, error) {
	llmResUncasted, ok := chunk.Metadata["llm_res"]
	if !ok {
		return "", []uint64{}, fmt.Errorf("llm_res metadata value not found")
	}

	llmRes, ok := llmResUncasted.(string)
	if !ok {
		return "", []uint64{}, fmt.Errorf("llm_res metadata value not of string type")
	}

	referenceIDsUncasted, ok := chunk.Metadata["reference_ids"]
	if !ok {
		return "", []uint64{}, fmt.Errorf("llm_res metadata value not found")
	}

	referenceIdsString, ok := referenceIDsUncasted.(string)
	if !ok {
		return "", []uint64{}, fmt.Errorf("llm_res metadata value not of string type")
	}

	referenceIds, err := referenceIdsFromString(referenceIdsString)
	if err != nil {
		return "", []uint64{}, err
	}

	return llmRes, referenceIds, nil
}

func (c *LLMCache) Query(query string, expectedReferenceIds []uint64) (string, error) {
	slog.Info("executing cache request", "query", query)

	chunks, err := c.Ndb.Query(query, 5, nil)
	if err != nil {
		return "", fmt.Errorf("ndb query error: %v", err)
	}

	if len(chunks) == 0 {
		slog.Info("cache returned no results")
		return "", nil
	}

	topChunk, topSimilarity := getTopChunkSimilarity(query, chunks)

	if topSimilarity < c.Threshold {
		slog.Info("top chunk similarity below threshold", "similarity", topSimilarity, "threshold", c.Threshold)
		return "", nil
	}

	llmRes, actualReferenceIds, err := getChunkMetadata(topChunk)
	if err != nil {
		return "", fmt.Errorf("error reading cache chunk metadata: %v", err)
	}

	// if the references match from the original query stored in the cache, we
	// can be pretty certain the llm response is still valid, thus return it
	slices.Sort(expectedReferenceIds)
	slices.Sort(actualReferenceIds)
	if slices.Equal(expectedReferenceIds, actualReferenceIds) {
		return llmRes, nil
	}

	// if the references have changed for the same query, delete it from the cache
	// since the underlying neuraldb has changed and the response might not be valid
	if query == topChunk.Text {
		err := c.Ndb.Delete(query, false)
		if err != nil {
			return "", fmt.Errorf("failed to delete cache entry: %v", err)
		}
	}

	// since the underlying references have changed, we don't return any response here
	return "", nil
}

func (c *LLMCache) Insert(query, llmRes string, referenceIds []uint64) error {
	slog.Info("inserting to cache", "query", query, "llm_res", llmRes)

	err := c.Ndb.Insert(
		"cache_query", query, // use the query as the docId so we can easily delete
		[]string{query},
		[]map[string]interface{}{{"llm_res": llmRes, "reference_ids": referenceIdsToString(referenceIds)}},
		nil)

	if err != nil {
		return fmt.Errorf("failed insertion to cache")
	}

	return nil
}
