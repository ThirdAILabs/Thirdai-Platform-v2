package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/nomad"
)

func StartLlmDispatchJob(client nomad.NomadClient, driver nomad.Driver, modelBazaarEndpoint, shareDir string) error {
	slog.Info("starting llm-dispatch job")

	job := nomad.LlmDispatchJob{
		ModelBazaarEndpoint: modelBazaarEndpoint,
		Driver:              driver,
		ShareDir:            shareDir,
	}

	if driver.DriverType() == "local" {
		// When running in production with docker we don't restart here because multiple
		// model bazaar jobs could be used. When an installation is updated the docker
		// version will be updated which will cause nomad to detect a change in the hcl
		// file and thus restart the job when StartJob is invoked later. If multiple
		// model bazaar jobs call StartJob with the same version, nomad will ignore
		// subsequent calls.
		err := nomad.StopJobIfExists(client, job.GetJobName())
		if err != nil {
			slog.Error("error stopping existing llm-dispatch job", "error", err)
			return fmt.Errorf("error stopping existing llm-dispatch job: %w", err)
		}
	}

	err := client.StartJob(job)
	if err != nil {
		slog.Error("error starting llm-dispatch job", "error", err)
		return fmt.Errorf("error starting llm-dispatch job: %w", err)
	}

	slog.Info("llm-dispatch job started successfully")
	return nil
}
