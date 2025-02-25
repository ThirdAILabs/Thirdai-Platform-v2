package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/orchestrator"
)

func StartLlmDispatchJob(orchestratorClient orchestrator.Client, driver orchestrator.Driver, modelBazaarEndpoint, shareDir string) error {
	slog.Info("starting llm-dispatch job")

	job := orchestrator.LlmDispatchJob{
		ModelBazaarEndpoint: modelBazaarEndpoint,
		Driver:              driver,
		ShareDir:            shareDir,
		IngressHostname:     orchestratorClient.IngressHostname(),
	}

	if driver.DriverType() == "local" {
		// When running in production with docker we don't restart here because multiple
		// model bazaar jobs could be used. When an installation is updated the docker
		// version will be updated which will cause nomad to detect a change in the hcl
		// file and thus restart the job when StartJob is invoked later. If multiple
		// model bazaar jobs call StartJob with the same version, nomad will ignore
		// subsequent calls.
		err := orchestrator.StopJobIfExists(orchestratorClient, job.GetJobName())
		if err != nil {
			slog.Error("error stopping existing llm-dispatch job", "error", err)
			return fmt.Errorf("error stopping existing llm-dispatch job: %w", err)
		}
	}

	err := orchestratorClient.StartJob(job)
	if err != nil {
		slog.Error("error starting llm-dispatch job", "error", err)
		return fmt.Errorf("error starting llm-dispatch job: %w", err)
	}

	slog.Info("llm-dispatch job started successfully")
	return nil
}
