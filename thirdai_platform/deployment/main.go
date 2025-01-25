package deployment

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	config, err := config.LoadConfig(env.ConfigPath)
	if err != nil {
		log.Fatalf("could not reads deployment config: %v", err.Error())
	}

	logFile, err := os.OpenFile(filepath.Join(config.ModelBazaarDir, "logs/", config.ModelId.String(), "deployment.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	initLogging(logFile)

	reporter := Reporter{config.ModelBazaarEndpoint, config.ModelId.String(), env.JobToken}

	ndbrouter, err := NewNdbRouter(config, reporter)
	if err != nil {
		log.Fatalf("failed to setup deployment router: %v", err)
	}
	defer ndbrouter.Close()
	
	r := ndbrouter.Routes()

	slog.Info("starting server", "port", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
	if err != nil {
		log.Fatalf("listen and serve returned error: %v", err.Error())
	}
}
