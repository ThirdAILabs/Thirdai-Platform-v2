package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/services"
)

func StartLlmCacheJob(client nomad.NomadClient, license *licensing.LicenseVerifier, vars *services.Variables) error {
	slog.Info("starting llm-cache job")

	licenseKey, err := license.Verify(0)
	if err != nil {
		slog.Error("license verification failed for llm-cache job", "error", err)
		return fmt.Errorf("license verification failed for llm-cache job: %w", err)
	}

	job := nomad.LlmCacheJob{
		ModelBazaarEndpoint: vars.ModelBazaarEndpoint,
		LicenseKey:          licenseKey.BoltLicenseKey,
		ShareDir:            vars.ShareDir,
		Driver:              vars.BackendDriver,
	}

	err = stopJobIfExists(client, job.GetJobName())
	if err != nil {
		slog.Error("error stopping existing llm-cache job", "error", err)
		return fmt.Errorf("error stopping existing llm-cache job: %w", err)
	}

	err = client.StartJob(job)
	if err != nil {
		slog.Error("error starting llm-cache job", "error", err)
		return fmt.Errorf("error starting llm-cache job: %w", err)
	}

	slog.Info("llm-cache job started successfully")
	return nil
}
