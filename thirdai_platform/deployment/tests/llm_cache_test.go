package tests

import (
	"testing"
	"thirdai_platform/deployment"

	"github.com/stretchr/testify/assert"
)

func TestLLMCache(t *testing.T) {
	err := verifyTestLicense()
	if err != nil {
		t.Fatalf("license error: %v", err)
	}

	modelbazaardir := t.TempDir()
	modelID := "test_model"

	cache, err := deployment.NewLLMCache(modelbazaardir, modelID)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	defer cache.Close()

	suggestions, err := cache.Suggestions("test query")
	assert.NoError(t, err)
	assert.Empty(t, suggestions)

	_, err = cache.Query("test query")
	assert.Error(t, err)

	err = cache.Insert("test query", "test response", []uint64{0, 1, 2})
	assert.NoError(t, err)

	suggestions, err = cache.Suggestions("test query")
	assert.NoError(t, err)
	assert.NotEmpty(t, suggestions)

	response, err := cache.Query("test query")
	assert.NoError(t, err)
	assert.Equal(t, "test response", response)

	response, err = cache.Query("test query and other diluting tokens")
	assert.NoError(t, err)
	assert.Equal(t, "", response)
}
