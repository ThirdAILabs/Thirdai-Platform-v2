package llm_dispatch

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
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
	client *openai.Client
}

func NewOpenAILLM(config LLMConfig) *OpenAILLM {
	client := openai.NewClient(config.APIKey)
	return &OpenAILLM{client: client}
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
	errChan := make(chan error)

	go func() {
		defer close(textChan)
		defer close(errChan)

		systemPrompt, userPrompt := makePrompt(req.Query, req.TaskPrompt, req.References)

		stream, err := l.client.CreateChatCompletionStream(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: req.Model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: userPrompt,
					},
				},
				Stream: true,
			},
		)
		if err != nil {
			errChan <- fmt.Errorf("error creating chat completion stream: %w", err)
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("error receiving from stream: %w", err)
				return
			}

			if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
				textChan <- response.Choices[0].Delta.Content
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

// LLMProvider defines the interface for LLM implementations
type LLMProvider interface {
	// Stream generates text from the LLM in a streaming fashion
	// Returns a channel for text chunks and a channel for errors
	// The text channel will be closed when streaming is complete
	// The error channel will be closed when streaming is complete or an error occurs
	Stream(req *GenerateRequest) (<-chan string, <-chan error)
} 

// This pattern helps in easily mocking the LLM provider in tests
// NewLLMProviderFunc is the type for the provider factory function
type NewLLMProviderFunc func(provider, apiKey string, logger *slog.Logger) (LLMProvider, error)

// NewLLMProvider creates a new LLM provider based on the specified type
var NewLLMProvider NewLLMProviderFunc = func(provider, apiKey string, logger *slog.Logger) (LLMProvider, error) {
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