package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/orchestrator"
)

func StartLlmCacheJob(orchestratorClient orchestrator.Client, license *licensing.LicenseVerifier, driver orchestrator.Driver, modelBazaarEndpoint, shareDir string) error {
	slog.Info("starting llm-cache job")

	licenseKey, err := license.Verify(0)
	if err != nil {
		slog.Error("license verification failed for llm-cache job", "error", err)
		return fmt.Errorf("license verification failed for llm-cache job: %w", err)
	}

	job := orchestrator.LlmCacheJob{
		ModelBazaarEndpoint: modelBazaarEndpoint,
		LicenseKey:          licenseKey.BoltLicenseKey,
		ShareDir:            shareDir,
		Driver:              driver,
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
			slog.Error("error stopping existing llm-cache job", "error", err)
			return fmt.Errorf("error stopping existing llm-cache job: %w", err)
		}
	}

	err = orchestratorClient.StartJob(job)
	if err != nil {
		slog.Error("error starting llm-cache job", "error", err)
		return fmt.Errorf("error starting llm-cache job: %w", err)
	}

	slog.Info("llm-cache job started successfully")
	return nil
}
