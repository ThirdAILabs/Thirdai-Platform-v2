package deployment

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/utils"
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

func main() {
	port := flag.Int("port", 8000, "Port to run server on")

	flag.Parse()

	env := loadEnv()

	config, err := config.LoadDeployConfig(env.ConfigPath)
	if err != nil {
		log.Fatalf("could not reads deployment config: %v", err.Error())
	}

	// TODO(david) should we inherit the ModelClient and implement new methods
	// for update status instead of adding them directly to ModelClient?
	reporter := client.NewModelClient(config.ModelBazaarEndpoint, env.JobToken, config.ModelId)

	logFile, err := os.OpenFile(filepath.Join(config.ModelBazaarDir, "logs/", config.ModelId.String(), "deployment.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		reporter.UpdateDeployStatusInternal("failed")
		log.Fatalf("error opening log file: %v", err)
	}
	//TODO (david) I think these defer calls don't run if we exit with log.Fatalf
	// we might need to move this code to a separate function that returns on error
	defer logFile.Close()

	initLogging(logFile)

	ndbrouter, err := NewNdbRouter(config, reporter)
	if err != nil {
		reporter.UpdateDeployStatusInternal("failed")
		log.Fatalf("failed to setup deployment router: %v", err)
	}
	defer ndbrouter.Close()

	r := ndbrouter.Routes()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		reporter.UpdateDeployStatusInternal("failed")
		log.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	slog.Info("server is ready to receive requests", "port", *port)
	reporter.UpdateDeployStatusInternal("complete")

	err = server.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		reporter.UpdateDeployStatusInternal("failed")
		log.Fatalf("listen and serve returned error: %v", err.Error())
	}
}
