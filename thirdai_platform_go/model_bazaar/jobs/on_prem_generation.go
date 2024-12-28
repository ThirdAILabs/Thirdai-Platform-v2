package jobs

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/storage"
)

func StartOnPremGenerationJobDefaultArgs(
	client nomad.NomadClient,
	storage storage.Storage,
	docker nomad.DockerEnv,
) error {
	return StartOnPremGenerationJob(client, storage, docker, "", true, true, -1)
}

func StartOnPremGenerationJob(
	client nomad.NomadClient,
	storage storage.Storage,
	docker nomad.DockerEnv,
	model string,
	restart bool,
	autoscaling bool,
	coresPerAllocation int,
) error {
	slog.Info("starting on-prem-generation job")
	if model == "" {
		model = "Llama-3.2-1B-Instruct-f16.gguf"
	}

	models, err := storage.List("gen-ai-models")
	if err != nil {
		slog.Error("error listing genai models", "error", err)
		return fmt.Errorf("error listing genai models: %w", err)
	}

	if !slices.Contains[[]string](models, model) {
		slog.Error("genai model not found", "model", model, "available_models", models)
		return fmt.Errorf("model %v is not available", model)
	}

	modelSize, err := storage.Size(filepath.Join("gen-ai-models", model))
	if err != nil {
		slog.Error("error getting model size", "error", err)
		return fmt.Errorf("error getting model size: %w", err)
	}

	if coresPerAllocation < 1 {
		coresPerAllocation = 7
	}

	job := nomad.OnPremLlmGenerationJob{
		AutoscalingEnabled: autoscaling,
		InitialAllocations: 1,
		MinAllocations:     1,
		MaxAllocations:     5,
		ModelDir:           filepath.Join(storage.Location(), "gen-ai-models"),
		ModelName:          model,
		Docker:             docker,
		Resources: nomad.Resources{
			AllocationMemory:    int(modelSize),
			AllocationMemoryMax: 2 * int(modelSize),
			AllocationCores:     coresPerAllocation,
		},
	}

	if !restart {
		exists, err := nomad.JobExists(client, job.GetJobName())
		if err != nil {
			slog.Error("error checking if on-prem-generation job exists", "error", err)
			return fmt.Errorf("error checking if on-prem-generation job exists: %w", err)
		}
		if exists {
			return nil
		}
	}

	err = nomad.StopJobIfExists(client, job.GetJobName())
	if err != nil {
		slog.Error("error stopping existing on-prem-generation job", "error", err)
		return fmt.Errorf("error stopping existing on-prem-generation job: %w", err)
	}

	err = client.StartJob(job)
	if err != nil {
		slog.Error("error starting on-prem-generation job", "error", err)
		return fmt.Errorf("error starting on-prem-generation job: %w", err)
	}

	slog.Info("on-prem-generation job started successfully")
	return nil

}
