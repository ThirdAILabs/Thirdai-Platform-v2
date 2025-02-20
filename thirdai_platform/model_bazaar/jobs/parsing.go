package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/nomad"
)

func StartParsingJob(client nomad.NomadClient, driver nomad.Driver, modelBazaarEndpoint, shareDir string) error {
	slog.Info("starting parsing job")

	job := nomad.ParsingJob{
		ModelBazaarEndpoint: modelBazaarEndpoint,
		ShareDir:            shareDir,
		Driver:              driver,
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
			slog.Error("error stopping existing parsing job", "error", err)
			return fmt.Errorf("error stopping existing parsing job: %w", err)
		}
	}

	err := client.StartJob(job)
	if err != nil {
		slog.Error("error starting parsing job", "error", err)
		return fmt.Errorf("error starting parsing job: %w", err)
	}

	slog.Info("parsing job started successfully")
	return nil
}
