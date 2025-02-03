package deployment

import (
	"fmt"
)

type GenAIClient interface {
	Stream() (<-chan string, <-chan error)
}

type OpenAILLM struct {
	GenAIClient
}

func (llm *OpenAILLM) Stream() (<-chan string, <-chan error) {
	textChan := make(chan string)
 	errChan := make(chan error, 1)

	return textChan, errChan
}

type OnPremLLM struct {
	GenAIClient
}

func (llm *OnPremLLM) Stream() (<-chan string, <-chan error) {
	textChan := make(chan string)
 	errChan := make(chan error, 1)

	return textChan, errChan
}

func GenAIClientFactory(provider string) (GenAIClient, error) {
	switch provider {
	case "openai":
		return &OpenAILLM{}, nil
	case "on-prem":
		return &OnPremLLM{}, nil
	default:
		return nil, fmt.Errorf("could not create LLM Client with provider %s", provider)
	}
}
