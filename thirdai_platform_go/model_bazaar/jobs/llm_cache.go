package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
)

func StartLlmCacheJob(client nomad.NomadClient, license *licensing.LicenseVerifier, driver nomad.Driver, modelBazaarEndpoint, shareDir string) error {
	slog.Info("starting llm-cache job")

	licenseKey, err := license.Verify(0)
	if err != nil {
		slog.Error("license verification failed for llm-cache job", "error", err)
		return fmt.Errorf("license verification failed for llm-cache job: %w", err)
	}

	job := nomad.LlmCacheJob{
		ModelBazaarEndpoint: modelBazaarEndpoint,
		LicenseKey:          licenseKey.BoltLicenseKey,
		ShareDir:            shareDir,
		Driver:              driver,
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
