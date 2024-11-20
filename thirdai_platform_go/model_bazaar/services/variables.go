package services

import (
	"fmt"
	"thirdai_platform/model_bazaar/nomad"
)

type Variables struct {
	Driver nomad.Driver

	ModelBazaarEndpoint string

	CloudCredentials nomad.CloudCredentials

	LlmProviders map[string]string
}

func (vars *Variables) GenaiKey(provider string) (string, error) {
	if provider == "on-prem" {
		return "", nil
	}
	key, ok := vars.LlmProviders[provider]
	if !ok {
		return "", fmt.Errorf("no api key specified for '%v', please set a key or use a different provider", provider)
	}
	return key, nil
}
