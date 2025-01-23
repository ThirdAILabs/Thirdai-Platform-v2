package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/jobs"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/services"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


type ndbDeploymentEnv struct {
	PublicModelBazaarEndpoint  string
	PrivateModelBazaarEndpoint string
	LicensePath                string
	NomadEndpoint              string
	NomadToken                 string
	ShareDir                   string
	JwtSecret                  string

	AdminUsername string
	AdminEmail    string
	AdminPassword string

	// LlmAutoscalingEnabled bool // TODO: is this needed
	GenAiKey string

	IdentityProvider      string
	KeycloakServerUrl     string
	UseSslInLogin         bool
	KeycloakAdminUsername string
	keycloakAdminPassword string

	MajorityCriticalServiceNodes int

	DockerRegistry string
	DockerUsername string
	DockerPassword string
	Tag            string
	BackendImage   string
	FrontendImage  string

	// These args are only needed if backend image is not specified, this is used to run locally.
	PythonPath  string
	PlatformDir string

	DatabaseUri  string
	GrafanaDbUri string

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
 func loadEnv() ndbDeploymentEnv {
	missingEnvs := []string{}

	requiredEnv := func(key string) string {
		env := os.Getenv(key)
		if env == "" {
			missingEnvs = append(missingEnvs, key)
			slog.Error("missing required env variable", "key", key)
		}
		return env
	}

	env := ndbDeploymentEnv{
		PublicModelBazaarEndpoint:  requiredEnv("PUBLIC_MODEL_BAZAAR_ENDPOINT"),
		PrivateModelBazaarEndpoint: requiredEnv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
		LicensePath:                requiredEnv("LICENSE_PATH"),
		NomadEndpoint:              requiredEnv("NOMAD_ENDPOINT"),
		NomadToken:                 requiredEnv("TASK_RUNNER_TOKEN"),
		ShareDir:                   requiredEnv("SHARE_DIR"),
		JwtSecret:                  requiredEnv("JWT_SECRET"),

		AdminUsername: requiredEnv("ADMIN_USERNAME"),
		AdminEmail:    requiredEnv("ADMIN_MAIL"),
		AdminPassword: requiredEnv("ADMIN_PASSWORD"),

		GenAiKey: optionalEnv("GENAI_KEY"),

		IdentityProvider:      requiredEnv("IDENTITY_PROVIDER"),
		KeycloakServerUrl:     optionalEnv("KEYCLOAK_SERVER_URL"),
		UseSslInLogin:         boolEnvVar("USE_SSL_IN_LOGIN"),
		KeycloakAdminUsername: optionalEnv("KEYCLOAK_ADMIN_USER"),
		keycloakAdminPassword: optionalEnv("KEYCLOAK_ADMIN_PASSWORD"),

		MajorityCriticalServiceNodes: intEnvVar("MAJORITY_CRITICAL_SERVICE_NODES", 1),

		DockerRegistry: requiredEnv("DOCKER_REGISTRY"),
		DockerUsername: requiredEnv("DOCKER_USERNAME"),
		DockerPassword: requiredEnv("DOCKER_PASSWORD"),
		Tag:            optionalEnv("TAG"),
		BackendImage:   optionalEnv("JOBS_IMAGE_NAME"),
		FrontendImage:  optionalEnv("FRONTEND_IMAGE_NAME"),

		PythonPath:  optionalEnv("PYTHON_PATH"),
		PlatformDir: optionalEnv("PLATFORM_DIR"),

		DatabaseUri:  requiredEnv("DATABASE_URI"),
		GrafanaDbUri: requiredEnv("GRAFANA_DB_URL"),

		CloudCredentials: nomad.CloudCredentials{
			AwsAccessKey:       optionalEnv("AWS_ACCESS_KEY"),
			AwsAccessSecret:    optionalEnv("AWS_ACCESS_SECRET"),
			AwsRegionName:      optionalEnv("AWS_REGION_NAME"),
			AzureAccountName:   optionalEnv("AZURE_ACCOUNT_NAME"),
			AzureAccountKey:    optionalEnv("AZURE_ACCOUNT_KEY"),
			GcpCredentialsFile: optionalEnv("GCP_CREDENTIALS_FILE"),
		},
	}

	if len(missingEnvs) > 0 {
		log.Fatalf("The following required env vars are missing: %s", strings.Join(missingEnvs, ", "))
	}

	if env.BackendImage == "" && (env.PythonPath == "" || env.PlatformDir == "") {
		log.Fatal("If JOBS_IMAGE_NAME env var is not specified then PYTHON_PATH and PLATFORM_DIR env vars must be provided.")
	} else if (env.BackendImage != "" || env.FrontendImage != "") && env.Tag == "" {
		log.Fatal("If JOBS_IMAGE_NAME or FRONTEND_IMAGE_NAME env vars are specified then TAG must be specified as well.")
	}

	return env
}


func main() {
	port := flag.Int("port", 8000, "Port to run server on")

	flag.Parse()

	env := loadEnv()

	r := chi.NewRouter()
	r.Mount("/api/v2", model_bazaar.Routes())

	slog.Info("starting server", "port", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
	if err != nil {
		log.Fatalf("listen and serve returned error: %v", err.Error())
	}
}