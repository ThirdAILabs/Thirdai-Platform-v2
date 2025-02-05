package tests

import (
	"testing"
	"thirdai_platform/deployment"
)

func TestLLMCache(t *testing.T) {
	err := verifyTestLicense()
	if err != nil {
		t.Fatalf("license error: %v", err)
	}

	modelbazaardir := t.TempDir()
	modelID := "test_model"

	cache, err := deployment.NewLLMCache(modelbazaardir, modelID)
	if err != nil {
		t.Fatalf("failed to create LLMCache: %v", err)
	}
	if cache == nil {
		t.Fatalf("cache should not be nil")
	}
	defer cache.Close()

	suggestions, err := cache.Suggestions("test query")
	if err != nil {
		t.Fatalf("failed to get suggestions: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("suggestions should be empty, got: %v", suggestions)
	}

	answer, err := cache.Query("test query", []uint64{0})
	if err != nil {
		t.Fatalf("error from cache.Query: %v", err)
	}
	if answer != "" {
		t.Fatalf("expected no results from query on empty cache, got answer %s", answer)
	}

	err = cache.Insert("test query", "test response", []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("failed to insert into cache: %v", err)
	}

	suggestions, err = cache.Suggestions("test query")
	if err != nil {
		t.Fatalf("failed to get suggestions: %v", err)
	}
	if len(suggestions) == 0 {
		t.Fatalf("suggestions should not be empty")
	}
	if suggestions[0] != "test query" {
		t.Fatalf("incorrect suggestion %s", suggestions[0])
	}

	response, err := cache.Query("test query", []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("failed to query cache: %v", err)
	}
	if response != "test response" {
		t.Fatalf("expected response 'test response', got '%s'", response)
	}

	response, err = cache.Query("test query and other diluting tokens", []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("failed to query cache: %v", err)
	}
	if response != "" {
		t.Fatalf("expected empty response, got '%s'", response)
	}
}
