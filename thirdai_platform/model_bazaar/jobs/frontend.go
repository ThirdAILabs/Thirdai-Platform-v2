package jobs

import (
	"fmt"
	"log/slog"
	"thirdai_platform/model_bazaar/orchestrator"
)

type FrontendJobArgs struct {
	IdentityProvider             string
	KeycloakServerHostname       string
	NextAuthSecret               string
	MajorityCriticalServiceNodes int
	UseSslInLogin                bool
	OpenaiKey                    string
	IsLocal                      bool
}

func StartFrontendJob(orchestratorClient orchestrator.Client, driver orchestrator.DockerDriver, args FrontendJobArgs) error {
	slog.Info("starting frontend job")

	job := orchestrator.FrontendJob{
		OpenaiApiKey:                 args.OpenaiKey,
		IdentityProvider:             args.IdentityProvider,
		KeycloakServerHostname:       args.KeycloakServerHostname,
		NextAuthSecret:               args.NextAuthSecret,
		MajorityCriticalServiceNodes: args.MajorityCriticalServiceNodes,
		UseSslInLogin:                args.UseSslInLogin,
		Driver:                       driver,
		IngressHostname:              orchestratorClient.IngressHostname(),
		IsLocal:                      args.IsLocal,
	}

	err := orchestratorClient.StartJob(job)
	if err != nil {
		slog.Error("error starting frontend job", "error", err)
		return fmt.Errorf("error starting frontend job: %w", err)
	}

	slog.Info("frontend job started successfully")
	return nil
}
