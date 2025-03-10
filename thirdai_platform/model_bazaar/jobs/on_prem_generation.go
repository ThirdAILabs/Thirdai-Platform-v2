package jobs

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/storage"
)

func StartOnPremGenerationJobDefaultArgs(
	orchestratorClient orchestrator.Client,
	storage storage.Storage,
	docker orchestrator.DockerEnv,
	isLocal bool,
) error {
	return StartOnPremGenerationJob(orchestratorClient, storage, docker, "", true, true, -1, isLocal)
}

const genaiModelsPath = "pretrained-models/genai"

func StartOnPremGenerationJob(
	orchestratorClient orchestrator.Client,
	storage storage.Storage,
	docker orchestrator.DockerEnv,
	model string,
	restart bool,
	autoscaling bool,
	coresPerAllocation int,
	isLocal bool,
) error {
	slog.Info("starting on-prem-generation job")
	if model == "" {
		model = "Llama-3.2-1B-Instruct-f16.gguf"
	}

	models, err := storage.List(genaiModelsPath)
	if err != nil {
		slog.Error("error listing genai models", "error", err)
		return fmt.Errorf("error listing genai models: %w", err)
	}

	if !slices.Contains[[]string](models, model) {
		slog.Error("genai model not found", "model", model, "available_models", models)
		return fmt.Errorf("model %v is not available", model)
	}

	modelSize, err := storage.Size(filepath.Join(genaiModelsPath, model))
	if err != nil {
		slog.Error("error getting model size", "error", err)
		return fmt.Errorf("error getting model size: %w", err)
	}

	if coresPerAllocation < 1 {
		coresPerAllocation = 7
	}

	job := orchestrator.OnPremLlmGenerationJob{
		AutoscalingEnabled: autoscaling,
		InitialAllocations: 1,
		MinAllocations:     1,
		MaxAllocations:     5,
		ModelDir:           filepath.Join(storage.Location(), genaiModelsPath),
		ModelName:          model,
		Docker:             docker,
		Resources: orchestrator.Resources{
			AllocationMemory:    int(modelSize),
			AllocationMemoryMax: 2 * int(modelSize),
			AllocationCores:     coresPerAllocation,
		},
		IngressHostname: orchestratorClient.IngressHostname(),
		IsLocal:         isLocal,
	}

	if !restart {
		exists, err := orchestrator.JobExists(orchestratorClient, job.GetJobName())
		if err != nil {
			slog.Error("error checking if on-prem-generation job exists", "error", err)
			return fmt.Errorf("error checking if on-prem-generation job exists: %w", err)
		}
		if exists {
			return nil
		}
	}

	err = orchestrator.StopJobIfExists(orchestratorClient, job.GetJobName())
	if err != nil {
		slog.Error("error stopping existing on-prem-generation job", "error", err)
		return fmt.Errorf("error stopping existing on-prem-generation job: %w", err)
	}

	err = orchestratorClient.StartJob(job)
	if err != nil {
		slog.Error("error starting on-prem-generation job", "error", err)
		return fmt.Errorf("error starting on-prem-generation job: %w", err)
	}

	slog.Info("on-prem-generation job started successfully")
	return nil

}
