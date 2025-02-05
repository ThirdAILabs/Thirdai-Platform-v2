package tests

import (
	"testing"
	"thirdai_platform/deployment"
)

func checkCacheQuery(t *testing.T, cache *deployment.LLMCache, query string, referenceIds []uint64, expectedAnswer string) {
	response, err := cache.Query(query, referenceIds)
	if err != nil {
		t.Fatalf("failed to query cache: %v", err)
	}
	if response != expectedAnswer {
		t.Fatalf("expected response %s, got '%s'", expectedAnswer, response)
	}
}

func checkCacheSuggestions(t *testing.T, cache *deployment.LLMCache, query string, expectedSuggestions []string) {
	suggestions, err := cache.Suggestions(query)
	if err != nil {
		t.Fatalf("failed to get cache suggestions: %v", err)
	}
	for i, expected := range expectedSuggestions {
		if suggestions[i] != expected {
			t.Fatalf("suggestion mismatch at index %d: expected %s, got %s", i, expected, suggestions[i])
		}
	}
}

func TestLLMCache(t *testing.T) {
	err := verifyTestLicense()
	if err != nil {
		t.Fatalf("license error: %v", err)
	}

	modelbazaardir := t.TempDir()
	modelID := "test_model"

	cache, err := deployment.NewLLMCache(modelbazaardir, modelID)
	if err != nil || cache == nil {
		t.Fatalf("failed to create LLMCache: %v", err)
	}
	defer cache.Close()

	checkCacheSuggestions(t, cache, "test query", []string{})

	checkCacheQuery(t, cache, "test query", []uint64{0}, "")

	err = cache.Insert("test query", "test response", []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("failed to insert into cache: %v", err)
	}

	checkCacheSuggestions(t, cache, "test query", []string{"test query"})
	checkCacheQuery(t, cache, "test query", []uint64{0, 1, 2}, "test response")
	checkCacheQuery(t, cache, "test query and other diluting tokens", []uint64{0, 1, 2}, "")

	// test eviction after incorrect ref ids are queried
	checkCacheQuery(t, cache, "test query", []uint64{100, 200, 300}, "")
	checkCacheQuery(t, cache, "test query", []uint64{0, 1, 2}, "")

	// multiple insertion shouldn't fail
	err = cache.Insert("test query", "test response", []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("failed to insert into cache: %v", err)
	}
	err = cache.Insert("test query", "another response", []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("failed to insert into cache: %v", err)
	}

	// test eviction kicks out all instances of a query
	checkCacheQuery(t, cache, "test query", []uint64{}, "")
	checkCacheQuery(t, cache, "test query", []uint64{0, 1, 2}, "")
}
