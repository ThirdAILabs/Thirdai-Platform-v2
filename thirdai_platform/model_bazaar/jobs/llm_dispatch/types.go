package llm_dispatch

import (
	"net/http"
	"time"
)

// Reference represents a piece of reference text with optional metadata
type Reference struct {
	Text     string                 `json:"text"`
	Source   string                 `json:"source,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateRequest represents an LLM generation request
type GenerateRequest struct {
	Query      string      `json:"query"`
	TaskPrompt string      `json:"task_prompt,omitempty"`
	References []Reference `json:"references,omitempty"`
	Key       string      `json:"key,omitempty"`
	Model     string      `json:"model"`
	Provider  string      `json:"provider"`
}

// LLMConfig holds configuration for an LLM provider
type LLMConfig struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// DefaultHTTPClient returns an http.Client with sensible defaults
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// LLMProvider defines the interface for LLM implementations
type LLMProvider interface {
	// Stream generates text from the LLM in a streaming fashion
	// Returns a channel for text chunks and a channel for errors
	// The text channel will be closed when streaming is complete
	// The error channel will be closed when streaming is complete or an error occurs
	Stream(req *GenerateRequest) (<-chan string, <-chan error)
} 