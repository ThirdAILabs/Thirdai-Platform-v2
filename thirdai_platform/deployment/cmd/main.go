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
	"syscall"
	"thirdai_platform/client"
	"thirdai_platform/deployment"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"time"

	"github.com/caarlos0/env/v10"
)

type CloudCredentials struct {
	AwsAccessKey       string `env:"AWS_ACCESS_KEY"`
	AwsAccessSecret    string `env:"AWS_ACCESS_SECRET"`
	AwsRegionName      string `env:"AWS_REGION_NAME"`
	AzureAccountName   string `env:"AZURE_ACCOUNT_NAME"`
	AzureAccountKey    string `env:"AZURE_ACCOUNT_KEY"`
	GcpCredentialsFile string `env:"GCP_CREDENTIALS_FILE"`
}

type DeploymentEnv struct {
	ConfigPath       string           `env:"CONFIG_PATH,required"`
	JobToken         string           `env:"JOB_TOKEN,required"`
	CloudCredentials CloudCredentials `env:""`
}

/**
 * ==========================================================================
 * ==== All variables used by the deployment job must be loaded here.    ====
 * ==== This is to make the data flow clear so that a user can see what  ====
 * ==== variables are exposed, and how the values are propagated through ====
 * ==== the system.                                                      ====
 * ==========================================================================
 */
func loadEnv() (*DeploymentEnv, error) {
	cfg := &DeploymentEnv{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
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

	env, err := loadEnv()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	config, err := config.LoadDeployConfig(env.ConfigPath)
	if err != nil {
		return fmt.Errorf("could not read deployment config: %w", err)
	}

	reporter := deployment.Reporter{BaseClient: client.NewBaseClient(config.ModelBazaarEndpoint, env.JobToken), ModelId: config.ModelId.String()}

	err = licensing.ActivateThirdAILicense(config.LicenseKey)
	if err != nil {
		return fmt.Errorf("could not activate thirdai license: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(config.ModelBazaarDir, "logs/", config.ModelId.String(), "deployment.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		reportErr := reporter.UpdateDeployStatusInternal("failed")
		if reportErr != nil {
			slog.Error("error reporting deploy status", "error", reportErr)
		}
		return fmt.Errorf("error opening log file: %w", err)
	}
	defer logFile.Close()

	initLogging(logFile)

	ndbrouter, err := deployment.NewNdbRouter(config, reporter)
	if err != nil {
		reportErr := reporter.UpdateDeployStatusInternal("failed")
		if reportErr != nil {
			slog.Error("error reporting deploy status", "error", reportErr)
		}
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
		reportErr := reporter.UpdateDeployStatusInternal("complete")
		if reportErr != nil {
			slog.Error("error reporting deploy status", "error", reportErr)
		} else {
			slog.Info("updated deploy status to complete")
		}
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
		reportErr := reporter.UpdateDeployStatusInternal("failed")
		if reportErr != nil {
			slog.Error("error reporting deploy status", "error", reportErr)
		}
		return fmt.Errorf("listen and serve returned error: %w", err)
	} else if err == http.ErrServerClosed {
		slog.Info("exited server with err=http.ErrServerClosed")
	}

	<-idleConnsClosed
	reportErr := reporter.UpdateDeployStatusInternal("stopped")
	if reportErr != nil {
		slog.Error("error reporting deploy status", "error", reportErr)
	} else {
		slog.Info("updated deploy status to stopped")
	}
	return nil
}

func main() {
	if err := runApp(); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}
