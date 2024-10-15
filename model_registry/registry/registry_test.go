package registry_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"model_registry/registry"
	"model_registry/schema"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	defaultEmail    = "admin@mail.com"
	defaultPassword = "password"
)

func setupRegistry(t *testing.T) *registry.ModelRegistry {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(&schema.Model{}, &schema.AccessCode{}, &schema.Admin{})
	if err != nil {
		t.Fatal(err)
	}

	localStoragePath := "./tmp/storage"

	storage := registry.NewLocalStorage(localStoragePath)

	registry := registry.New(db, storage)

	registry.AddAdmin(defaultEmail, defaultPassword)

	t.Cleanup(func() {
		err := os.RemoveAll(localStoragePath)
		if err != nil {
			t.Fatal(err)
		}
	})

	return registry
}

func login(router chi.Router, t *testing.T) string {
	body := fmt.Sprintf(`{"email": "%v", "password": "%v"}`, defaultEmail, defaultPassword)
	req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Login should succeed")
	}

	result := make(map[string]string)
	json.NewDecoder(resp.Body).Decode(&result)

	return result["token"]
}

func authHeader(token string) string {
	return fmt.Sprintf("Bearer %v", token)
}

func TestAdminLogin(t *testing.T) {
	registry := setupRegistry(t)
	router := registry.Routes()

	dummyParams := map[string]interface{}{
		"model_name":    "abc",
		"model_type":    "1",
		"model_subtype": "2",
		"access":        "public",
		"size":          1,
		"checksum":      "123",
	}

	body, err := json.Marshal(dummyParams)
	if err != nil {
		t.Fatal(err)
	}

	{
		req := httptest.NewRequest("POST", "/upload-start", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusUnauthorized {
			t.Fatal("Request should be unauthorized without token")
		}
	}

	token := login(router, t)

	{
		req := httptest.NewRequest("POST", "/upload-start", bytes.NewReader(body))
		req.Header.Add("Authorization", authHeader(token))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("Request should succeed with token %d %v", w.Result().StatusCode, w.Body.String())
		}
	}
}

func getChecksum(data []byte) string {
	dataChecksum, err := registry.Checksum(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return dataChecksum
}

func uploadModel(router chi.Router, name string, data []byte, checksum string, token string, access string) error {
	params := map[string]interface{}{
		"model_name":    name,
		"description":   "a model called " + name,
		"model_type":    name,
		"model_subtype": name + "::" + name,
		"access":        access,
		"metadata":      "",
		"size":          len(data),
		"checksum":      checksum,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req := httptest.NewRequest("POST", "/upload-start", bytes.NewReader(body))
	req.Header.Add("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload start failed %d, %v", resp.StatusCode, w.Body.String())
	}

	result := make(map[string]string)
	json.NewDecoder(resp.Body).Decode(&result)
	session := result["session_token"]

	for start := 0; start < len(data); start += 10 {
		end := min(start+10, len(data))

		req := httptest.NewRequest("POST", "/upload-chunk", bytes.NewReader(data[start:end]))
		req.Header.Add("Authorization", authHeader(session))
		req.Header.Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			return fmt.Errorf("upload chunk failed %d %v", w.Result().StatusCode, w.Body.String())
		}
	}

	req = httptest.NewRequest("POST", "/upload-commit", nil)
	req.Header.Add("Authorization", authHeader(session))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		return fmt.Errorf("upload commit failed %d, %v", w.Result().StatusCode, w.Body.String())
	}

	return nil
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

var unauthorizedError = errors.New("download link is unauthorized")

func getDownloadLink(router chi.Router, name string, accessCode string) (string, error) {
	params := map[string]interface{}{
		"model_name":  name,
		"access_code": accessCode,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return "", nil
	}
	req := httptest.NewRequest("POST", "/download-link", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	code := w.Result().StatusCode
	if code == http.StatusUnauthorized {
		return "", unauthorizedError
	}
	if code != http.StatusOK {
		return "", fmt.Errorf("get download link failed %d %v", code, w.Body.String())
	}

	result := make(map[string]string)
	json.NewDecoder(w.Body).Decode(&result)
	return result["download_link"], nil
}

func downloadModel(router chi.Router, name string, token string) ([]byte, error) {
	link, err := getDownloadLink(router, name, token)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest("GET", link, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed %d, %v", w.Result().StatusCode, w.Body.String())
	}

	return w.Body.Bytes(), nil
}

func createAccessCode(router chi.Router, modelName string, token string) (string, error) {
	params := map[string]interface{}{
		"model_name":       modelName,
		"access_code_name": modelName + "-code",
	}
	body, err := json.Marshal(params)
	if err != nil {
		return "", nil
	}

	req := httptest.NewRequest("POST", "/generate-access-code", bytes.NewReader(body))
	req.Header.Add("Authorization", authHeader(token))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	code := w.Result().StatusCode
	if code != http.StatusOK {
		return "", fmt.Errorf("Failed to generate token %d %v", code, w.Body.String())
	}

	result := make(map[string]string)
	json.NewDecoder(w.Body).Decode(&result)
	return result["access_code"], nil
}

func listModels(t *testing.T, router chi.Router, accessCodes []string) []registry.ModelInfo {
	params := map[string]interface{}{
		"access_codes": accessCodes,
	}
	body, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/list-models", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("list models failed %d %v", w.Result().StatusCode, w.Body.String())
	}

	result := make(map[string][]registry.ModelInfo)
	json.NewDecoder(w.Body).Decode(&result)
	return result["models"]
}

func TestFullModelWorkflow(t *testing.T) {
	registry := setupRegistry(t)
	router := registry.Routes()

	token := login(router, t)

	model1 := randomBytes(107)
	model2 := randomBytes(362)
	model3 := randomBytes(84)

	t.Run("UploadModels", func(t *testing.T) {
		err := uploadModel(router, "abc", model1, getChecksum(model1), token, schema.Public)
		if err != nil {
			t.Fatal(err)
		}
		err = uploadModel(router, "xyz", model2, getChecksum(model2), token, schema.Private)
		if err != nil {
			t.Fatal(err)
		}
		err = uploadModel(router, "123", model3, getChecksum(model3), token, schema.Private)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DownloadPublicModel", func(t *testing.T) {
		download, err := downloadModel(router, "abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(model1, download) {
			t.Fatal("downloaded model doesn't match")
		}
	})

	t.Run("DownloadPrivateModelFails", func(t *testing.T) {
		_, err := downloadModel(router, "xyz", "")
		if err != unauthorizedError {
			t.Fatal(err)
		}
	})

	var code1, code2 string

	t.Run("CreateAccessCodes", func(t *testing.T) {
		c1, err := createAccessCode(router, "abc", token)
		if err != nil {
			t.Fatal(err)
		}
		code1 = c1

		c2, err := createAccessCode(router, "xyz", token)
		if err != nil {
			t.Fatal(err)
		}
		code2 = c2
	})

	t.Run("DownloadPrivateModelFailsWrongToken", func(t *testing.T) {
		_, err := downloadModel(router, "xyz", code1)
		if err != unauthorizedError {
			t.Fatal(err)
		}
	})

	t.Run("DownloadPrivateModelFailsRandomTokenToken", func(t *testing.T) {
		_, err := downloadModel(router, "xyz", "lsjdflk")
		if err != unauthorizedError {
			t.Fatal(err)
		}
	})

	t.Run("DownloadPrivateModelWithToken", func(t *testing.T) {
		download, err := downloadModel(router, "xyz", code2)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(model2, download) {
			t.Fatal("downloaded model doesn't match")
		}
	})

	t.Run("ListModelsWithoutAccessCode", func(t *testing.T) {
		models := listModels(t, router, []string{})
		if len(models) != 1 {
			t.Fatal("Expected 1 model")
		}

		if models[0].Name != "abc" {
			t.Fatal("incorrect model returned")
		}
	})

	t.Run("ListModelsWithWrongAccessCode", func(t *testing.T) {
		models := listModels(t, router, []string{code1})
		if len(models) != 1 {
			t.Fatal("Expected 1 model")
		}

		if models[0].Name != "abc" {
			t.Fatal("incorrect model returned")
		}
	})

	t.Run("ListModelsWithAccessCode", func(t *testing.T) {
		models := listModels(t, router, []string{code1, code2})
		if len(models) != 2 {
			t.Fatal("Expected 2 models")
		}

		if models[0].Name != "abc" || models[1].Name != "xyz" {
			t.Fatal("incorrect model returned")
		}
	})

	t.Run("DeleteModel", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/delete-model?model_name=xyz", nil)
		req.Header.Add("Authorization", authHeader(token))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("failed to delete model %d %v", w.Result().StatusCode, w.Body.String())
		}

		if len(listModels(t, router, []string{code1, code2})) != 1 {
			t.Fatalf("Expected only 1 result after deleting model")
		}
	})
}

func TestModelChecksums(t *testing.T) {
	registry := setupRegistry(t)
	router := registry.Routes()

	token := login(router, t)

	model := randomBytes(528)
	checksum := getChecksum(model)
	model[48] = model[48] + 1 // Corrupt byte

	err := uploadModel(router, "abc", model, checksum, token, schema.Public)
	if err == nil {
		t.Fatal("Checksum mismatch should cause upload to fail")
	}

	if !strings.Contains(err.Error(), "Checksum doesn't match for model") {
		t.Fatal("Expected checksum error")
	}
}
