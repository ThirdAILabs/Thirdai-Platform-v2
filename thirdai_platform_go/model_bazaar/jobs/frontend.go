package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/services"
)

func StartFrontendJob(client nomad.NomadClient, vars *services.Variables) error {
	slog.Info("starting frontend job")

	openaiKey, err := vars.GenaiKey("openai")
	if err != nil {
		openaiKey = ""
	}
	job := nomad.FrontendJob{
		OpenaiApiKey:           openaiKey,
		IdentityProvider:       "TODO",
		KeycloakServerHostname: "TODO",
		NextAuthSecret:         "TODO",
		UseSslInLogin:          false, // TODO
		Driver:                 vars.FrontendDriver,
	}

	err = stopJobIfExists(client, job.GetJobName())
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
