package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"thirdai_platform/client"
	"thirdai_platform/deployment"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/utils"
	"time"
)

type deploymentEnv struct {
	ConfigPath string
	JobToken   string

	CloudCredentials nomad.CloudCredentials
}

/**
 * ==========================================================================
 * ==== All variables used by the deployment job must be loaded here.    ====
 * ==== This is to make the data flow clear so that a user can see what  ====
 * ==== variables are exposed, and how the values are propagated through ====
 * ==== the system.                                                      ====
 * ==========================================================================
 */
func loadEnv() deploymentEnv {
	missingEnvs := []string{}

	requiredEnv := func(key string) string {
		env := os.Getenv(key)
		if env == "" {
			missingEnvs = append(missingEnvs, key)
			slog.Error("missing required env variable", "key", key)
		}
		return env
	}

	env := deploymentEnv{
		ConfigPath: requiredEnv("CONFIG_PATH"),
		JobToken:   requiredEnv("JOB_TOKEN"),

		CloudCredentials: nomad.CloudCredentials{
			AwsAccessKey:       utils.OptionalEnv("AWS_ACCESS_KEY"),
			AwsAccessSecret:    utils.OptionalEnv("AWS_ACCESS_SECRET"),
			AwsRegionName:      utils.OptionalEnv("AWS_REGION_NAME"),
			AzureAccountName:   utils.OptionalEnv("AZURE_ACCOUNT_NAME"),
			AzureAccountKey:    utils.OptionalEnv("AZURE_ACCOUNT_KEY"),
			GcpCredentialsFile: utils.OptionalEnv("GCP_CREDENTIALS_FILE"),
		},
	}

	if len(missingEnvs) > 0 {
		log.Fatalf("The following required env vars are missing: %s", strings.Join(missingEnvs, ", "))
	}

	return env
}

func initLogging(logFile *os.File) {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	slog.Info("logging initialized", "log_file", logFile.Name())
}

// The reason we have a separate runApp function is because the defer calls don't
// run if we exit with log.Fatalf, so instead we return an err here and fail outside
func runApp() error {
	port := flag.Int("port", 8000, "Port to run server on")

	flag.Parse()

	env := loadEnv()

	config, err := config.LoadDeployConfig(env.ConfigPath)
	if err != nil {
		return fmt.Errorf("could not read deployment config: %w", err)
	}

	licensing.ActivateThirdAILicense(config.LicenseKey)

	reporter := deployment.Reporter{BaseClient: client.NewBaseClient(config.ModelBazaarEndpoint, env.JobToken), ModelId: config.ModelId.String()}

	logFile, err := os.OpenFile(filepath.Join(config.ModelBazaarDir, "logs/", config.ModelId.String(), "deployment.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		reporter.UpdateDeployStatusInternal("failed")
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer logFile.Close()

	initLogging(logFile)

	ndbrouter, err := deployment.NewNdbRouter(config, reporter)
	if err != nil {
		reporter.UpdateDeployStatusInternal("failed")
		return fmt.Errorf("failed to setup deployment router: %w", err)
	}
	defer ndbrouter.Close()

	r := ndbrouter.Routes()

	/* If we report the server is complete before traefik updates, a user might
	fire a request to this deployment before traefik is ready, and that request
	will fail. Since traefik updates are every 5 seconds, this should be a safeguard.
	The caveat is any code that comes after this should not take more than 10s. */
	go func() {
		time.Sleep(10 * time.Second)
		reporter.UpdateDeployStatusInternal("complete")
		slog.Info("updated deploy status to complete")
	}()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}

	/* We need to listen for an interrupt in this way to ensure the defer calls
	go through correctly in case of a shutdown and so we can update the job
	status to "stopped"
	srv.Shutdown shuts down the server without interrupting valid connections
	https://pkg.go.dev/net/http#Server.Shutdown */
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutdown signal received")
		if err := srv.Shutdown(context.Background()); err != nil {
			slog.Error("HTTP server Shutdown", "err", err)
		}
		close(idleConnsClosed)
	}()

	slog.Info("starting server", "port", *port)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		reporter.UpdateDeployStatusInternal("failed")
		return fmt.Errorf("listen and serve returned error: %w", err)
	} else if err == http.ErrServerClosed {
		slog.Info("exited server with err=http.ErrServerClosed")
	}

	<-idleConnsClosed
	reporter.UpdateDeployStatusInternal("stopped")
	slog.Info("updated deploy status to stopped")
	return nil
}

func main() {
	if err := runApp(); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}
