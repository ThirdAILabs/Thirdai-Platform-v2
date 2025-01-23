package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/utils"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RecoveryService struct {
	db       *gorm.DB
	storage  storage.Storage
	nomad    nomad.NomadClient
	userAuth auth.IdentityProvider

	variables Variables
}

func (s *RecoveryService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(s.userAuth.AuthMiddleware()...)
	r.Use(auth.AdminOnly(s.db))

	r.Post("/backup", s.Backup)
	r.Get("/backups", s.ListLocalBackups)

	return r
}

func (s *RecoveryService) getDbUri() (string, error) {
	postgres, ok := s.db.Dialector.(*postgres.Dialector)
	if !ok {
		return "", fmt.Errorf("snapshot is only available for postgres db")
	}

	fields := make(map[string]string)
	for _, field := range strings.Split(postgres.DSN, " ") {
		parts := strings.Split(field, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid field in DSN: %v", field)
		}
		fields[parts[0]] = parts[1]
	}

	for _, expected := range []string{"user", "password", "host", "port", "dbname"} {
		if _, ok := fields[expected]; !ok {
			return "", fmt.Errorf("expected field '%v' in DSN", expected)
		}
	}

	return fmt.Sprintf("postgresql://%v:%v@%v:%v/%v", fields["user"], fields["password"], fields["host"], fields["port"], fields["dbname"]), nil
}

func (s *RecoveryService) saveConfig(params BackupRequest, configPath string) error {
	providerInfo := map[string]string{"provider": params.Provider}
	if params.Provider != "local" {
		providerInfo["bucket_name"] = params.BucketName
	}
	switch params.Provider {
	case "s3":
		providerInfo["aws_access_key"] = s.variables.CloudCredentials.AwsAccessKey
		providerInfo["aws_secret_access_key"] = s.variables.CloudCredentials.AwsAccessSecret
	case "azure":
		providerInfo["azure_account_name"] = s.variables.CloudCredentials.AzureAccountName
		providerInfo["azure_account_key"] = s.variables.CloudCredentials.AzureAccountKey
	case "gcp":
		providerInfo["gcp_credentials_file_path"] = s.variables.CloudCredentials.GcpCredentialsFile
	case "local":
	// pass
	default:
		return fmt.Errorf("invalid provider: '%v'", params.Provider)
	}

	config := map[string]interface{}{
		"provider": providerInfo, "interval_minutes": params.IntervalMinutes, "backup_limit": params.BackupLimit,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error serializing snapshot config: %w", err)
	}

	err = s.storage.Write(configPath, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error saving snapshot config: %w", err)
	}

	return nil
}

type BackupRequest struct {
	Provider   string `json:"provider"`
	BucketName string `json:"bucket_name"`

	IntervalMinutes *int `json:"interval_minutes"`
	BackupLimit     *int `json:"backup_limit"`
}

func (s *RecoveryService) Backup(w http.ResponseWriter, r *http.Request) {
	var params BackupRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	dbUri, err := s.getDbUri()
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to perform backup, unable to validate DB connection string: %v", err), http.StatusBadRequest)
		return
	}

	configPath := "backup_config.json"
	err = s.saveConfig(params, configPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job := nomad.SnapshotJob{
		// TODO(Any): this is needed because the snapshot job does not use the storage interface
		// in the future once this is standardized it will not be needed
		ConfigPath: filepath.Join(s.storage.Location(), configPath),
		ShareDir:   s.variables.ShareDir,
		DbUri:      dbUri,
		Driver:     s.variables.BackendDriver,
	}

	err = nomad.StopJobIfExists(s.nomad, job.GetJobName())
	if err != nil {
		http.Error(w, fmt.Sprintf("error stopping existing snapshot job: %v", err), http.StatusBadRequest)
		return
	}

	err = s.nomad.StartJob(job)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting snapshot job: %v", err), http.StatusBadRequest)
		return
	}

	utils.WriteSuccess(w)
}

func (s *RecoveryService) ListLocalBackups(w http.ResponseWriter, r *http.Request) {
	exists, err := s.storage.Exists("backups")
	if err != nil {
		http.Error(w, fmt.Sprintf("error checking if local backups exist: %v", err), http.StatusBadRequest)
		return
	}

	if !exists {
		utils.WriteJsonResponse(w, []string{})
		return
	}

	backups, err := s.storage.List("backups")
	if err != nil {
		http.Error(w, fmt.Sprintf("error listing local backups: %v", err), http.StatusBadRequest)
		return
	}

	utils.WriteJsonResponse(w, backups)
}
