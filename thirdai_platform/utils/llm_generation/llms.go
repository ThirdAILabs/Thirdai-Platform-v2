package llm_generation

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type OpenAILLM struct {
	client *openai.Client
}

func NewOpenAILLM(apiKey string) *OpenAILLM {
	client := openai.NewClient(apiKey)
	return &OpenAILLM{client: client}
}

type OnPremLLM struct {
	endpoint string
	client   *http.Client
}

func NewOnPremLLM() (*OnPremLLM, error) {
	model_bazaar_endpoint := os.Getenv("MODEL_BAZAAR_ENDPOINT")
	if model_bazaar_endpoint == "" {
		return nil, fmt.Errorf("MODEL_BAZAAR_ENDPOINT not set")
	}

	baseURL, err := url.JoinPath(model_bazaar_endpoint, "on-prem-llm/v1/chat/completions")
	if err != nil {
		return nil, fmt.Errorf("error creating API URL: %w", err)
	}

	return &OnPremLLM{
		endpoint: baseURL,
		client:   DefaultHTTPClient(),
	}, nil
}

func (l *OnPremLLM) makeRequest(method string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, l.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := l.client.Do(req)
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

func makePrompt(query, inputTaskPrompt string, refs []Reference) (string, string) {
	var refTexts []string
	for _, ref := range refs {
		if ext := strings.ToLower(filepath.Ext(ref.Source)); ext == ".pdf" || ext == ".docx" || ext == ".csv" {
			refTexts = append(refTexts, fmt.Sprintf(`(From file "%s") %s`, ref.Source, ref.Text))
		} else {
			refTexts = append(refTexts, fmt.Sprintf(`(From a webpage) %s`, ref.Text))
		}
	}

	context := strings.Join(refTexts, "\n\n")

	tokenLimit := 2000
	words := strings.Fields(context)
	if len(words) > tokenLimit {
		context = strings.Join(words[:tokenLimit], " ")
	}

	const defaultSystemPrompt = "Write a short answer for the user's query based on the provided context. " +
		"If the context provides insufficient information, mention it but answer to " +
		"the best of your abilities."

	const defaultTaskPrompt = "Given this context, "

	systemPrompt := defaultSystemPrompt
	taskPrompt := defaultTaskPrompt
	if inputTaskPrompt != "" {
		taskPrompt = inputTaskPrompt
	}
	userPrompt := fmt.Sprintf("%s\n\n %s %s", context, taskPrompt, query)

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
			"stream":    true,
			"n_predict": 1000,
			"model":     req.Model,
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			errChan <- fmt.Errorf("error marshaling request: %w", err)
			return
		}

		resp, err := l.makeRequest("POST", jsonBody, map[string]string{})
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
type NewLLMProviderFunc func(provider, apiKey string) (LLMProvider, error)

var NewLLMProvider NewLLMProviderFunc = func(provider, apiKey string) (LLMProvider, error) {
	switch strings.ToLower(provider) {
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("API key required for OpenAI")
		}
		return NewOpenAILLM(apiKey), nil
	case "on-prem":
		return NewOnPremLLM()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func StreamResponse(llm LLMProvider, req *GenerateRequest, w http.ResponseWriter, r *http.Request) (string, error) {
	textChan, errChan := llm.Stream(req)

	w.Header().Set("Content-Type", "text/event-stream")
	flusher, ok := w.(http.Flusher)
	if !ok {
		return "", fmt.Errorf("streaming unsupported")
	}

	var accumulatedResponse bytes.Buffer

	for {
		select {
		case text, ok := <-textChan:
			if !ok {
				return accumulatedResponse.String(), nil
			}
			fmt.Fprintf(w, "data: %s\n\n", text)
			flusher.Flush()

			accumulatedResponse.WriteString(text)

		case err, ok := <-errChan:
			if !ok {
				return accumulatedResponse.String(), nil
			}
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
			flusher.Flush()
			return accumulatedResponse.String(), err

		case <-r.Context().Done():
			return accumulatedResponse.String(), nil
		}
	}
}
