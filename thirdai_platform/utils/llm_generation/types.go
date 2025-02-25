package llm_generation

import (
	"net/http"
	"time"
)

type Reference struct {
	Id     uint64 `json:"reference_id"`
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

type GenerateRequest struct {
	Query      string      `json:"query"`
	TaskPrompt string      `json:"task_prompt"`
	References []Reference `json:"references,omitempty"`
	Model      string      `json:"model"`
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
