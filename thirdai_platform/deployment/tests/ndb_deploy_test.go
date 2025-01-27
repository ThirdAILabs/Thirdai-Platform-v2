package tests

import (
	"slices"
	"testing"
	"thirdai_platform/model_bazaar/utils"
	"thirdai_platform/deployment"
	"thirdai_platform/ndb"
	"time"
)

func TestDeployConfig(t *testing.T) {
	utils.SaveConfig()
}


func TestQueryHandler(t *testing.T) {
	router := deployment.NdbRouter{
		ndb: mockNdb{ // Mock NeuralDB implementation
			QueryFn: func(query string, topk int) ([]ndb.Chunk, error) {
				return []ndb.Chunk{{Id: 1, Text: "Result", Document: "Source", Score: 0.9}}, nil
			},
		},
	}

	req := httptest.NewRequest("POST", "/query", strings.NewReader(`{"query":"test","top_k":1}`))
	w := httptest.NewRecorder()

	router.Search(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var results searchResults
	json.NewDecoder(resp.Body).Decode(&results)

	require.Len(t, results.References, 1)
	require.Equal(t, "Result", results.References[0].Text)
}
