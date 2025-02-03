package llm_generation

import (
	"net/http"
	"time"
)

// Reference represents a piece of reference text
type Reference struct {
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

// GenerateRequest represents an LLM generation request
type GenerateRequest struct {
	Query      string      `json:"query"`
	TaskPrompt string      `json:"task_prompt,omitempty"`
	References []Reference `json:"references,omitempty"`
	Model      string      `json:"model"` // ex : gpt-4o-mini, claude-3-5-sonnet-20240620
}

// DefaultHTTPClient returns an http.Client with sensible defaults for connection pooling
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