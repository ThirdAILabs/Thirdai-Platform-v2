package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/nomad"
)

func StartFrontendJob(client nomad.NomadClient, driver nomad.DockerDriver, openaiKey string) error {
	slog.Info("starting frontend job")

	job := nomad.FrontendJob{
		OpenaiApiKey:                 openaiKey,
		IdentityProvider:             "TODO",
		KeycloakServerHostname:       "TODO",
		NextAuthSecret:               "TODO",
		MajorityCriticalServiceNodes: 1,     // TODO
		UseSslInLogin:                false, // TODO
		Driver:                       driver,
	}

	err := stopJobIfExists(client, job.GetJobName())
	if err != nil {
		slog.Error("error stopping existing frontend job", "error", err)
		return fmt.Errorf("error stopping existing frontend job: %w", err)
	}

	err = client.StartJob(job)
	if err != nil {
		slog.Error("error starting frontend job", "error", err)
		return fmt.Errorf("error starting frontend job: %w", err)
	}

	slog.Info("frontend job started successfully")
	return nil
}
