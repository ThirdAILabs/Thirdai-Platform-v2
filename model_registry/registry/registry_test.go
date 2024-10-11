package registry_test

import (
	"encoding/json"
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
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.AutoMigrate(&schema.Model{}, &schema.AccessToken{}, &schema.Admin{})
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

func TestAdminLogin(t *testing.T) {
	registry := setupRegistry(t)
	router := registry.Routes()

	{
		req := httptest.NewRequest("POST", "/upload-start", strings.NewReader(`{"model_name": "abc"}`))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Result().StatusCode != http.StatusUnauthorized {
			t.Fatal("Request should be unauthorized without token")
		}
	}

	token := login(router, t)

	{
		req := httptest.NewRequest("POST", "/upload-start", strings.NewReader(`{"model_name": "abc"}`))
		req.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Fatal("Request should succeed with token")
		}
	}
}
