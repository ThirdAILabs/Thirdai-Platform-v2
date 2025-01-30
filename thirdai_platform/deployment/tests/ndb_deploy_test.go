package tests

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"testing"

	"thirdai_platform/deployment"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/search/ndb"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func makeNdbServer(t *testing.T) (*httptest.Server, error) {
	modelbazaardir := t.TempDir()
	modelID := uuid.New()
	modelDir := filepath.Join(modelbazaardir, "models", modelID.String(), "model", "model.ndb")

	db, err := ndb.New(modelDir)
	if err != nil {
		return nil, err
	}

	err = db.Insert(
		"doc_name_1", "doc_id_1",
		[]string{"test line one", "another test line", "something without that"},
		[]map[string]interface{}{{"thing1": true}, {"thing2": true}, {"thing1": true}},
		nil,
	)
	if err != nil {
		return nil, err
	}

	slog.Info("NDB initialized successfully")

	deployConfig := config.DeployConfig{
		ModelId:        modelID,
		ModelBazaarDir: modelbazaardir,
	}
	router := deployment.NdbRouter{Ndb: db, Config: &deployConfig}

	r := router.Routes()
	testServer := httptest.NewServer(r)

	return testServer, nil
}

func checkHealth(t *testing.T, testServer *httptest.Server) {
	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func checkSources(t *testing.T, testServer *httptest.Server, sources []string) {
	resp, err := http.Get(testServer.URL + "/sources")
	if err != nil {
		t.Fatalf("GET /sources failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var data struct {
		Sources []struct {
			Source   string `json:"source"`
			SourceID string `json:"source_id"`
			Version  uint32 `json:"version"`
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode /query response: %v", err)
	}

	assert.True(t, len(data.Sources) != 0)
	for i, expected := range sources {
		assert.Equal(t, expected, data.Sources[i].SourceID, "source mismatch at index %d", i)
	}
}

// TODO(david) split this test into multiple unit tests
// TODO(david) reuse the tests from the integration tests and use the NDBClient? api is slightly different now though
func TestBasicEndpoints(t *testing.T) {
	testServer, err := makeNdbServer(t)
	if err != nil {
		t.Fatalf("Failed to make ndb server: %v", err)
	}
	defer testServer.Close()

	checkHealth(t, testServer)

	t.Run("Query", func(t *testing.T) {
		body := map[string]interface{}{
			"query": "test line",
			"top_k": 2,
		}
		bodyBytes, _ := json.Marshal(body)
		resp, err := http.Post(testServer.URL+"/query", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("POST /query failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var data struct {
			References []struct {
				ID     int    `json:"id"`
				Text   string `json:"text"`
				Source string `json:"source"`
				Score  float32
			} `json:"references"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatalf("failed to decode /query response: %v", err)
		}

		assert.True(t, len(data.References) != 0)

		reference_ids := make([]int, len(data.References))
		for i, ref := range data.References {
			reference_ids[i] = ref.ID
		}

		assert.True(t, slices.Contains(reference_ids, 0))
		assert.True(t, slices.Contains(reference_ids, 1))
	})

	t.Run("Insert", func(t *testing.T) {
		body := map[string]interface{}{
			"document": "doc_name_2",
			"doc_id":   "doc_id_2",
			"chunks":   []string{"a new word", "another new word"},
			"metadata": []map[string]interface{}{
				{"example": true},
				{"another": "test"},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		resp, err := http.Post(testServer.URL+"/insert", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("POST /insert failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		// TODO When you make a query for "word" you receive references 3 and 4
	})

	t.Run("Delete", func(t *testing.T) {
		body := map[string]interface{}{
			"source_ids": []string{"doc_id_1"},
		}
		bodyBytes, _ := json.Marshal(body)
		resp, err := http.Post(testServer.URL+"/delete", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("POST /delete failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Upvote", func(t *testing.T) {
		body := map[string]interface{}{
			"text_id_pairs": []map[string]interface{}{
				{
					"query_text":     "haha",
					"reference_id":   0,
					"reference_text": "test line one",
				},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		resp, err := http.Post(testServer.URL+"/upvote", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("POST /upvote failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Associate", func(t *testing.T) {
		body := map[string]interface{}{
			"text_pairs": []map[string]interface{}{
				{"source": "chunk A", "target": "test line one"},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		resp, err := http.Post(testServer.URL+"/associate", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("POST /associate failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	checkSources(t, testServer, []string{"doc_id_2"})
}
