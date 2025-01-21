package llm_dispatch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// baseLLM provides common functionality for all LLM providers
type baseLLM struct {
	config LLMConfig
}

func (b *baseLLM) makeRequest(method, endpoint string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := b.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// OpenAILLM implements LLMProvider for OpenAI
type OpenAILLM struct {
	baseLLM
}

func NewOpenAILLM(config LLMConfig) *OpenAILLM {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1/chat/completions"
	}
	if config.HTTPClient == nil {
		config.HTTPClient = DefaultHTTPClient()
	}
	return &OpenAILLM{baseLLM{config}}
}

// OnPremLLM implements LLMProvider for on-premises deployment
type OnPremLLM struct {
	baseLLM
}

func NewOnPremLLM(config LLMConfig) (*OnPremLLM, error) {
	endpoint := os.Getenv("MODEL_BAZAAR_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("MODEL_BAZAAR_ENDPOINT not set")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = DefaultHTTPClient()
	}

	baseURL, err := url.JoinPath(endpoint, "on-prem-llm/v1/chat/completions")
	if err != nil {
		return nil, fmt.Errorf("error creating API URL: %w", err)
	}
	config.BaseURL = baseURL

	return &OnPremLLM{baseLLM{config}}, nil
}

func makePrompt(query, taskPrompt string, refs []Reference) (string, string) {
	var systemPrompt string
	if len(refs) > 0 {
		var refTexts []string
		for _, ref := range refs {
			refTexts = append(refTexts, ref.Text)
		}
		systemPrompt = fmt.Sprintf("Use the following references to answer the question:\n%s", strings.Join(refTexts, "\n\n"))
	} else {
		systemPrompt = "You are a helpful AI assistant."
	}

	userPrompt := query
	if taskPrompt != "" {
		userPrompt = fmt.Sprintf("%s\n\n%s", taskPrompt, query)
	}

	return systemPrompt, userPrompt
}

func (l *OpenAILLM) Stream(req *GenerateRequest) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		systemPrompt, userPrompt := makePrompt(req.Query, req.TaskPrompt, req.References)

		body := map[string]interface{}{
			"model": req.Model,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userPrompt},
			},
			"stream": true,
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			errChan <- fmt.Errorf("error marshaling request: %w", err)
			return
		}

		resp, err := l.makeRequest("POST", l.config.BaseURL, jsonBody, map[string]string{
			"Authorization": "Bearer " + l.config.APIKey,
		})
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("error reading stream: %w", err)
				return
			}

			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				return
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content != "" {
							textChan <- content
						}
					}
				}
			}
		}
	}()

	return textChan, errChan
}

func (l *OnPremLLM) Stream(req *GenerateRequest) (<-chan string, <-chan error) {
	textChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(textChan)
		defer close(errChan)

		systemPrompt, userPrompt := makePrompt(req.Query, req.TaskPrompt, req.References)

		body := map[string]interface{}{
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": userPrompt},
			},
			"stream": true,
			"n_predict": 1000,
			"model": req.Model,
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			errChan <- fmt.Errorf("error marshaling request: %w", err)
			return
		}

		resp, err := l.makeRequest("POST", l.config.BaseURL, jsonBody, map[string]string{})
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("error reading stream: %w", err)
				return
			}

			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				return
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content != "" {
							textChan <- content
						}
					}
				}
			}
		}
	}()

	return textChan, errChan
}

// NewLLMProvider creates a new LLM provider based on the specified type
func NewLLMProvider(provider, apiKey string, logger *slog.Logger) (LLMProvider, error) {
	config := LLMConfig{
		APIKey:     apiKey,
		HTTPClient: DefaultHTTPClient(),
	}

	switch strings.ToLower(provider) {
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("API key required for OpenAI")
		}
		return NewOpenAILLM(config), nil
	case "on-prem":
		return NewOnPremLLM(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
} 