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

type modelBazaarEnv struct {
	PublicModelBazaarEndpoint  string
	PrivateModelBazaarEndpoint string
	LicensePath                string
	NomadEndpoint              string
	NomadToken                 string
	ShareDir                   string

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

func boolEnvVar(key string) bool {
	value := os.Getenv(key)
	return strings.ToLower(value) == "true"
}

func requiredEnv(key string) string {
	env := os.Getenv(key)
	if env == "" {
		log.Fatalf("Missing required env variable %v", key)
	}
	return env
}

func optionalEnv(key string) string {
	return os.Getenv(key)
}

func loadEnvFile(envFile string) {
	slog.Info(fmt.Sprintf("loading env from file %v", envFile))
	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("error loading .env file '%v': %v", envFile, err)
	}
}

/**
 * ==========================================================================
 * ==== All variables that are used by model bazaar must be loaded here. ====
 * ==== This is to make the data flow clear so that a user can see what  ====
 * ==== variables are exposed, and how the values are propagated through ====
 * ==== the system.                                                      ====
 * ==========================================================================
 */
func loadEnv() modelBazaarEnv {
	env := modelBazaarEnv{
		PublicModelBazaarEndpoint:  requiredEnv("PUBLIC_MODEL_BAZAAR_ENDPOINT"),
		PrivateModelBazaarEndpoint: requiredEnv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
		LicensePath:                requiredEnv("LICENSE_PATH"),
		NomadEndpoint:              requiredEnv("NOMAD_ENDPOINT"),
		NomadToken:                 requiredEnv("TASK_RUNNER_TOKEN"),
		ShareDir:                   requiredEnv("SHARE_DIR"),

		AdminUsername: requiredEnv("ADMIN_USERNAME"),
		AdminEmail:    requiredEnv("ADMIN_MAIL"),
		AdminPassword: requiredEnv("ADMIN_PASSWORD"),

		GenAiKey: optionalEnv("GENAI_KEY"),

		IdentityProvider:      requiredEnv("IDENTITY_PROVIDER"),
		KeycloakServerUrl:     optionalEnv("KEYCLOAK_SERVER_URL"),
		UseSslInLogin:         boolEnvVar("USE_SSL_IN_LOGIN"),
		KeycloakAdminUsername: optionalEnv("KEYCLOAK_ADMIN_USER"),
		keycloakAdminPassword: optionalEnv("KEYCLOAK_ADMIN_PASSWORD"),

		DockerRegistry: requiredEnv("DOCKER_REGISTRY"),
		DockerUsername: requiredEnv("DOCKER_USERNAME"),
		DockerPassword: requiredEnv("DOCKER_PASSWORD"),
		Tag:            optionalEnv("TAG"),
		BackendImage:   optionalEnv("THIRDAI_PLATFORM_IMAGE_NAME"),
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

	if env.BackendImage == "" && (env.PythonPath == "" || env.PlatformDir == "") {
		log.Fatal("If THIRDAI_PLATFORM_IMAGE_NAME env var is not specified then PYTHON_PATH and PLATFORM_DIR env vars must be provided.")
	} else if (env.BackendImage != "" || env.FrontendImage != "") && env.Tag == "" {
		log.Fatal("If THIRDAI_PLATFORM_IMAGE_NAME or FRONTEND_IMAGE_NAME env vars are specified then TAG must be specified as well.")
	}

	return env
}

func (env *modelBazaarEnv) postgresDsn() string {
	parts, err := url.Parse(env.DatabaseUri)
	if err != nil {
		log.Fatalf("error parsing db uri: %v", err)
	}
	pwd, _ := parts.User.Password()
	dbname := strings.TrimPrefix(parts.Path, "/")
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", parts.Hostname(), parts.User.Username(), pwd, dbname, parts.Port())
}

func (env *modelBazaarEnv) BackendDriver() nomad.Driver {
	if env.BackendImage == "" {
		return nomad.LocalDriver{
			PythonPath:  env.PythonPath,
			PlatformDir: env.PlatformDir,
		}
	}

	return nomad.DockerDriver{
		ImageName: env.BackendImage,
		Tag:       env.Tag,
		DockerEnv: nomad.DockerEnv{
			Registry:       env.DockerRegistry,
			DockerUsername: env.DockerUsername,
			DockerPassword: env.DockerPassword,
			ShareDir:       env.ShareDir,
		},
	}
}

func (env *modelBazaarEnv) FrontendDriver() nomad.DockerDriver {
	return nomad.DockerDriver{
		ImageName: env.FrontendImage,
		Tag:       env.Tag,
		DockerEnv: nomad.DockerEnv{
			Registry:       env.DockerRegistry,
			DockerUsername: env.DockerUsername,
			DockerPassword: env.DockerPassword,
			ShareDir:       env.ShareDir,
		},
	}
}

func (env *modelBazaarEnv) llmProviders() map[string]string {
	providers := map[string]string{}
	if strings.HasPrefix(env.GenAiKey, "sk-") {
		providers["openai"] = env.GenAiKey
	}
	return providers
}

func initLogging(logFile *os.File) {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)
	log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	slog.Info("logging initialized", "log_file", logFile.Name())
}

func initDb(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	err = db.AutoMigrate(
		&schema.Model{}, &schema.ModelAttribute{}, &schema.ModelDependency{},
		&schema.User{}, &schema.Team{}, &schema.UserTeam{}, &schema.JobLog{},
	)
	if err != nil {
		log.Fatalf("error migrating db schema: %v", err)
	}

	fmt.Println("DB: ", db.Dialector.(*postgres.Dialector).DSN)

	return db
}

func getHostname(u string) string {
	parts, err := url.Parse(u)
	if err != nil {
		log.Fatalf("error parsing url '%v': %v", u, err)
	}
	return parts.Hostname()
}

func main() {
	envFile := flag.String("env", "", "File to load env variables from")
	skipJobRestart := flag.Bool("skip_job_restart", false, "If true will not restart llm-cache, llm-dispatch, and telemetry jobs")
	port := flag.Int("port", 8000, "Port to run server on")

	flag.Parse()

	if *envFile != "" {
		loadEnvFile(*envFile)
	}
	env := loadEnv()

	err := os.MkdirAll(filepath.Join(env.ShareDir, "logs/"), 0777)
	if err != nil {
		log.Fatalf("error creating log dir: %v", err)
	}

	logFile, err := os.OpenFile(filepath.Join(env.ShareDir, "logs/model_bazaar.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	auditLog, err := os.OpenFile(filepath.Join(env.ShareDir, "logs/audit.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening audit log file: %v", err)
	}
	defer auditLog.Close()

	initLogging(logFile)

	db := initDb(env.postgresDsn())

	nomadClient := nomad.NewHttpClient(env.NomadEndpoint, env.NomadToken)

	licenseVerifier := licensing.NewVerifier(env.LicensePath)

	sharedStorage := storage.NewSharedDisk(env.ShareDir)

	variables := services.Variables{
		BackendDriver: env.BackendDriver(),
		DockerRegistry: services.DockerRegistry{
			Registry:       env.DockerRegistry,
			DockerUsername: env.DockerUsername,
			DockerPassword: env.DockerPassword,
		},
		ShareDir:            env.ShareDir,
		ModelBazaarEndpoint: env.PrivateModelBazaarEndpoint,
		CloudCredentials:    env.CloudCredentials,
		LlmProviders:        env.llmProviders(),
	}

	var identityProvider auth.IdentityProvider
	if env.IdentityProvider == "keycloak" {
		identityProvider, err = auth.NewKeycloakIdentityProvider(
			db,
			auth.NewAuditLogger(auditLog),
			auth.KeycloakArgs{
				KeycloakServerUrl:     env.KeycloakServerUrl,
				KeycloakAdminUsername: env.KeycloakAdminUsername,
				KeycloakAdminPassword: env.keycloakAdminPassword,
				AdminUsername:         env.AdminUsername,
				AdminEmail:            env.AdminEmail,
				AdminPassword:         env.AdminPassword,
				PublicHostname:        getHostname(env.PublicModelBazaarEndpoint),
				PrivateHostname:       getHostname(env.PrivateModelBazaarEndpoint),
				SslLogin:              env.UseSslInLogin,
			},
		)
		if err != nil {
			log.Fatalf("error creating keycloak identity provider: %v", err)
		}
	} else {
		identityProvider, err = auth.NewBasicIdentityProvider(
			db,
			auth.NewAuditLogger(auditLog),
			auth.BasicProviderArgs{
				AdminUsername: env.AdminUsername,
				AdminEmail:    env.AdminEmail,
				AdminPassword: env.AdminPassword,
			},
		)
		if err != nil {
			log.Fatalf("error creating basic identity provider: %v", err)
		}
	}

	model_bazaar := services.NewModelBazaar(
		db,
		nomadClient,
		sharedStorage,
		licenseVerifier,
		identityProvider,
		variables,
	)

	if !*skipJobRestart {
		err = jobs.StartLlmCacheJob(nomadClient, licenseVerifier, env.BackendDriver(), env.PrivateModelBazaarEndpoint, env.ShareDir)
		if err != nil {
			log.Fatalf("failed to start llm cache job: %v", err)
		}

		err = jobs.StartLlmDispatchJob(nomadClient, env.BackendDriver(), env.PrivateModelBazaarEndpoint, env.ShareDir)
		if err != nil {
			log.Fatalf("failed to start llm dispatch job: %v", err)
		}

		telemetryArgs := jobs.TelemetryJobArgs{
			IsLocal:             env.BackendImage == "",
			ModelBazaarEndpoint: env.PrivateModelBazaarEndpoint,
			Docker:              variables.DockerEnv(),
			GrafanaDbUrl:        env.GrafanaDbUri,
			AdminUsername:       env.AdminUsername,
			AdminEmail:          env.AdminEmail,
			AdminPassword:       env.AdminPassword,
		}
		err = jobs.StartTelemetryJob(nomadClient, sharedStorage, telemetryArgs)
		if err != nil {
			log.Fatalf("failed to start telemetry job: %v", err)
		}
	}

	if env.FrontendImage != "" {
		err := jobs.StartFrontendJob(nomadClient, env.FrontendDriver(), env.GenAiKey)
		if err != nil {
			log.Fatalf("failed to start frontend job: %v", err)
		}
	}

	go model_bazaar.JobStatusSync(5 * time.Second)

	r := chi.NewRouter()
	r.Mount("/api/v2", model_bazaar.Routes())

	slog.Info("starting server", "port", *port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
	if err != nil {
		log.Fatalf("listen and serve returned error: %v", err.Error())
	}
	model_bazaar.StopJobStatusSync()
}
