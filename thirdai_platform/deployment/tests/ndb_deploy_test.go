package tests

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"thirdai_platform/deployment"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/services"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/search/ndb"
	"thirdai_platform/utils/llm_generation"

	"bufio"
	"strings"

	"github.com/google/uuid"
)

type MockLLM struct{}

func (m *MockLLM) Stream(req *llm_generation.GenerateRequest) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(textChan)
		defer close(errChan)

		textChan <- "This "
		textChan <- "is "
		textChan <- "a test."
	}()

	return textChan, errChan
}

type MockPermissions struct {
	GetModelPermissionsFunc   func(string) (services.ModelPermissions, error)
	ModelPermissionsCheckFunc func(string) func(http.Handler) http.Handler
	History                   map[string]int
}

func (m *MockPermissions) GetModelPermissions(token string) (services.ModelPermissions, error) {
	return services.ModelPermissions{Read: true, Write: true}, nil // Grant all permissions
}

func (m *MockPermissions) ModelPermissionsCheck(permission_type string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

func makeNdbServer(t *testing.T, modelbazaardir string) *httptest.Server {
	modelID := uuid.New()
	modelDir := filepath.Join(modelbazaardir, "models", modelID.String(), "model", "model.ndb")

	db, err := ndb.New(modelDir)
	if err != nil {
		t.Fatalf("failed to create NDB: %v", err)
	}

	err = db.Insert(
		"doc_name_1", "doc_id_1",
		[]string{"test line one", "another test line", "something without that"},
		[]map[string]interface{}{{"thing1": true}, {"thing2": true}, {"thing1": true}},
		nil,
	)
	if err != nil {
		t.Fatalf("failed to insert into NDB: %v", err)
	}

	slog.Info("NDB initialized successfully")

	deployConfig := config.DeployConfig{
		ModelId:        modelID,
		ModelBazaarDir: modelbazaardir,
	}

	mockPermissions := MockPermissions{}

	router := deployment.NdbRouter{Ndb: db, Config: &deployConfig, Permissions: &mockPermissions}

	r := router.Routes()
	testServer := httptest.NewServer(r)
	router.LLMProvider = &MockLLM{}

	return testServer
}

func checkHealth(t *testing.T, testServer *httptest.Server) {
	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("failed to get /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func checkSources(t *testing.T, testServer *httptest.Server, sources []string) {
	resp, err := http.Get(testServer.URL + "/sources")
	if err != nil {
		t.Fatalf("failed to get /sources: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var data deployment.Sources
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode /sources response: %v", err)
	}

	if len(data.Sources) == 0 {
		t.Fatalf("expected sources, but got none")
	}

	for i, expected := range sources {
		if data.Sources[i].SourceID != expected {
			t.Fatalf("source mismatch at index %d: expected %s, got %s", i, expected, data.Sources[i].SourceID)
		}
	}
}

func checkQuery(t *testing.T, testServer *httptest.Server, query string, reference_ids []int) {
	body := map[string]interface{}{
		"query": query,
		"top_k": 2,
	}
	bodyBytes, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+"/query", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to post /query: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var data deployment.SearchResults
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode /query response: %v", err)
	}

	if len(data.References) == 0 {
		t.Fatalf("expected references, but got none")
	}

	returned_reference_ids := make([]int, len(data.References))
	for i, ref := range data.References {
		returned_reference_ids[i] = ref.Id
	}

	for _, id := range reference_ids {
		if !slices.Contains(returned_reference_ids, id) {
			t.Fatalf("expected reference ID %d not found", id)
		}
	}
}
func doUpvote(t *testing.T, testServer *httptest.Server, query string, reference_id int) {
	body := map[string]interface{}{
		"text_id_pairs": []map[string]interface{}{
			{
				"query_text":     query,
				"reference_id":   reference_id,
				"reference_text": "test line one",
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+"/upvote", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func doAssociate(t *testing.T, testServer *httptest.Server, source string, target string) {
	body := map[string]interface{}{
		"text_pairs": []map[string]interface{}{
			{"source": source, "target": target},
		},
	}
	bodyBytes, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+"/associate", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func doInsert(t *testing.T, testServer *httptest.Server) {
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
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func doDelete(t *testing.T, testServer *httptest.Server, source_ids []string) {
	body := map[string]interface{}{
		"source_ids": source_ids,
	}
	bodyBytes, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+"/delete", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func doGenerate(t *testing.T, testServer *httptest.Server, query string, references []map[string]interface{}, model string) {
	body := map[string]interface{}{
		"query":       query,
		"task_prompt": "say your name",
		"references":  references,
		"model":       model,
	}

	bodyBytes, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+"/generate-with-references", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("Expected Content-Type 'text/event-stream', got %s", resp.Header.Get("Content-Type"))
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			fullResponse.WriteString(data)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Scanner encountered an error: %v", err)
	}
	if fullResponse.String() != "This is a test." {
		t.Fatalf("Expected response 'This is a test.', got %s", fullResponse.String())
	}
}

func TestBasicEndpoints(t *testing.T) {
	v := licensing.NewVerifier("platform_test_license.json")
	license, err := v.LoadLicense()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	licensing.ActivateThirdAILicense(license.License.BoltLicenseKey)

	modelbazaardir := t.TempDir()
	testServer := makeNdbServer(t, modelbazaardir)
	defer testServer.Close()

	checkSources(t, testServer, []string{"doc_id_1"})
	checkHealth(t, testServer)
	checkQuery(t, testServer, "test line", []int{0, 1})

	doAssociate(t, testServer, "source", "test line")
	checkQuery(t, testServer, "source", []int{0, 1})

	doUpvote(t, testServer, "unrelated query", 2)
	checkQuery(t, testServer, "unrelated query", []int{2})

	doInsert(t, testServer)
	checkSources(t, testServer, []string{"doc_id_1", "doc_id_2"})
	doDelete(t, testServer, []string{"doc_id_1"})
	checkSources(t, testServer, []string{"doc_id_2"})

	doGenerate(t, testServer, "is this a test?", []map[string]interface{}{
		{"text": "my name is chatgpt", "source": "doc_id_1"},
	}, "gpt-4o-mini")
}

func TestSaveLoadDeployConfig(t *testing.T) {
	expectedConfig := &config.DeployConfig{
		ModelId:             uuid.New(),
		UserId:              uuid.New(),
		ModelType:           "test-model",
		ModelBazaarDir:      "/bazaar/model",
		HostDir:             "/host/model",
		ModelBazaarEndpoint: "http://localhost:8080",
		LicenseKey:          "test-license-key",
		JobAuthToken:        "test-auth-token",
		Autoscaling:         true,
		Options: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}

	tmp_dir := t.TempDir()
	store := storage.NewSharedDisk(tmp_dir)

	configData, err := json.MarshalIndent(expectedConfig, "", "    ")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = store.Write("deploy_config.json", bytes.NewReader(configData))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	loadedConfig, err := config.LoadDeployConfig(filepath.Join(tmp_dir, "deploy_config.json"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expectedConfig, loadedConfig) {
		t.Fatalf("Loaded config should match the expected config")
	}
}

// TODO unit tests for constraints, full source paths, insertion of large files
