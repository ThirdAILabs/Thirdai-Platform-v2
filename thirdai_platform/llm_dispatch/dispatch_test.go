package llm_dispatch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockLLM implements LLMProvider interface for testing
type MockLLM struct{}

func (m *MockLLM) Stream(req *GenerateRequest) (<-chan string, <-chan error) {
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

func TestGenerateTextStream(t *testing.T) {
	tests := []struct {
		name       string
		references []string
		prompt     string
	}{
		{
			name:       "No references, no prompt",
			references: nil,
			prompt:     "",
		},
		{
			name:       "With references",
			references: []string{"Text from doc A", "Text from doc B"},
			prompt:     "",
		},
		{
			name:       "With prompt",
			references: nil,
			prompt:     "This is a custom prompt",
		},
	}

	// Store original provider factory
	originalProvider := NewLLMProvider
	// Restore it after tests
	defer func() { NewLLMProvider = originalProvider }()

	// Replace provider factory with mock
	NewLLMProvider = func(provider, apiKey string, logger *slog.Logger) (LLMProvider, error) {
		return &MockLLM{}, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			var refs []Reference
			for _, text := range tt.references {
				refs = append(refs, Reference{Text: text})
			}

			request := GenerateRequest{
				Query:      "test query",
				TaskPrompt: tt.prompt,
				References: refs,
				Provider:   "openai",
				Key:       "dummy key",
			}

			body, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/llm-dispatch/generate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}

			// Read and combine all streamed responses
			var fullResponse strings.Builder
			reader := bufio.NewReader(w.Body)
			for {
				line, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("Error reading response: %v", err)
				}
				fullResponse.WriteString(line)
			}

			// Check the complete response
			responseText := fullResponse.String()
			expected := "This is a test."
			
			// Extract actual text from SSE format
			var actualText strings.Builder
			for _, line := range strings.Split(responseText, "\n") {
				if strings.HasPrefix(line, "data: ") {
					actualText.WriteString(strings.TrimPrefix(line, "data: "))
				}
			}

			if actualText.String() != expected {
				t.Errorf("Expected response text %q, got %q", expected, actualText.String())
			}
		})
	}
}

func TestMissingAPIKey(t *testing.T) {
	router := NewRouter()

	request := map[string]interface{}{
		"query":    "test query",
		"provider": "openai",
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/llm-dispatch/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expected := "No generative AI key provided"
	if response["detail"] != expected {
		t.Errorf("Expected error message %q, got %q", expected, response["detail"])
	}
}

func TestUnsupportedProvider(t *testing.T) {
	router := NewRouter()

	request := map[string]interface{}{
		"query":    "test query",
		"provider": "unknown_provider",
		"key":      "dummy key",
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/llm-dispatch/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expected := "Unsupported provider: unknown_provider"
	if response["detail"] != expected {
		t.Errorf("Expected error message %q, got %q", expected, response["detail"])
	}
}

func TestInvalidRequestBody(t *testing.T) {
	router := NewRouter()

	// Missing required 'query' field
	request := map[string]interface{}{
		"provider": "openai",
		"key":      "dummy key",
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/llm-dispatch/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expected := "Field required: query"
	if response["detail"] != expected {
		t.Errorf("Expected error message %q, got %q", expected, response["detail"])
	}
}

func TestHealthCheck(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest("GET", "/llm-dispatch/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %q", response["status"])
	}
}