package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/services"
)

func StartLlmDispatchJob(client nomad.NomadClient, vars *services.Variables) error {
	slog.Info("starting llm-dispatch job")

	job := nomad.LlmDispatchJob{
		ModelBazaarEndpoint: vars.ModelBazaarEndpoint,
		Driver:              vars.BackendDriver,
	}

	err := stopJobIfExists(client, job.GetJobName())
	if err != nil {
		slog.Error("error stopping existing llm-dispatch job", "error", err)
		return fmt.Errorf("error stopping existing llm-dispatch job: %w", err)
	}

	err = client.StartJob(job)
	if err != nil {
		slog.Error("error starting llm-dispatch job", "error", err)
		return fmt.Errorf("error starting llm-dispatch job: %w", err)
	}

	slog.Info("llm-dispatch job started successfully")
	return nil
}
