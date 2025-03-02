package services

import (
	"fmt"
	"thirdai_platform/model_bazaar/orchestrator"
)

type DockerRegistry struct {
	Registry       string
	DockerUsername string
	DockerPassword string
}

type Variables struct {
	BackendDriver orchestrator.Driver

	DockerRegistry DockerRegistry

	ShareDir            string
	ModelBazaarEndpoint string

	CloudCredentials orchestrator.CloudCredentials

	LlmProviders map[string]string
	IsLocal      bool
}

func (vars *Variables) DockerEnv() orchestrator.DockerEnv {
	return orchestrator.DockerEnv{
		Registry:       vars.DockerRegistry.Registry,
		DockerUsername: vars.DockerRegistry.DockerUsername,
		DockerPassword: vars.DockerRegistry.DockerPassword,
		ShareDir:       vars.ShareDir,
	}
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
