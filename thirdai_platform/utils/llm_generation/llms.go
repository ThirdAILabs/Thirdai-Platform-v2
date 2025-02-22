package llm_generation

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLM interface {
	StreamResponse(req GenerateRequest, w http.ResponseWriter, r *http.Request) error
}

type LLMProvider string

const (
	OpenAILLM LLMProvider = "openai"
	OnPremLLM LLMProvider = "on-prem"
)

type OpenAICompliantLLM struct {
	client *openai.Client
}

func createOpenAILLMClient(apiKey string, endpoint *string) (*openai.Client, error) {
	var client *openai.Client

	if endpoint == nil {
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
		)
	} else {
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(*endpoint),
		)
	}
	return client, nil
}

func newOpenAILLM(apiKey string, endpoint *string) (LLM, error) {
	client, err := createOpenAILLMClient(apiKey, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenAI client: %w", err)
	}
	return &OpenAICompliantLLM{client: client}, nil
}

func newOnPremLLM() (LLM, error) {
	// assumes that onprem llm has openai compliant api
	model_bazaar_endpoint := os.Getenv("MODEL_BAZAAR_ENDPOINT")
	if model_bazaar_endpoint == "" {
		slog.Error("MODEL_BAZAAR_ENDPOINT not set")
		return nil, fmt.Errorf("MODEL_BAZAAR_ENDPOINT not set")
	}

	baseURL, err := url.JoinPath(model_bazaar_endpoint, "v1/")
	if err != nil {
		slog.Error("error creating API URL: %v")
		return nil, fmt.Errorf("error creating API URL: %w", err)
	}

	client, err := createOpenAILLMClient(
		"", // assumes that the onprem llm requires no api key
		&baseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating OpenAI client: %w", err)
	}

	return &OpenAICompliantLLM{
		client: client,
	}, nil
}

func NewLLM(provider LLMProvider, apiKey string) (LLM, error) {
	switch provider {
	case OpenAILLM:
		return newOpenAILLM(apiKey, nil)
	case OnPremLLM:
		return newOnPremLLM()
	default:
		slog.Error("invalid provider", "provider", provider)
		return nil, fmt.Errorf("invalid provider: %s", provider)
	}
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

func (llm *OpenAICompliantLLM) StreamResponse(req GenerateRequest, w http.ResponseWriter, r *http.Request) error {

	w.Header().Set("Content-Type", "text/event-stream")
	flusher, ok := w.(http.Flusher)
	if !ok {
		slog.Error("streaming unsupported")
		return fmt.Errorf("streaming unsupported")
	}

	systemPrompt, userPrompt := makePrompt(req.Query, req.TaskPrompt, req.References)

	messages := openai.F([]openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(userPrompt),
	})

	stream := llm.client.Chat.Completions.NewStreaming(
		context.Background(),
		openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    openai.F(req.Model),
		},
	)
	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			fmt.Fprintf(w, "data: %s\n\n", evt.Choices[0].Delta.Content)
			flusher.Flush()
		}
	}
	if err := stream.Err(); err != nil {
		slog.Error("error streaming response: %v", slog.String("error", err.Error()))
		return fmt.Errorf("error streaming response: %w", err)
	}
	return nil
}
