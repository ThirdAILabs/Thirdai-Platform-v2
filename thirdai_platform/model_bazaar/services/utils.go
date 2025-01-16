package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/model_bazaar/utils"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func listModelDependencies(modelId uuid.UUID, db *gorm.DB) ([]schema.Model, error) {
	visited := map[uuid.UUID]struct{}{}
	models := []schema.Model{}

	var recurse func(uuid.UUID) error

	recurse = func(m uuid.UUID) error {
		if _, ok := visited[m]; ok {
			return nil
		}

		visited[m] = struct{}{}

		model, err := schema.GetModel(m, db, true, false, false)
		if err != nil {
			return fmt.Errorf("error while listing model dependencies: %w", err)
		}

		for _, dep := range model.Dependencies {
			err := recurse(dep.DependencyId)
			if err != nil {
				return err
			}
		}

		models = append(models, model)
		return nil
	}

	err := recurse(modelId)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func getStatus(model *schema.Model, trainStatus bool) string {
	if trainStatus {
		return model.TrainStatus
	}
	return model.DeployStatus
}

func getModelStatus(model schema.Model, db *gorm.DB, trainStatus bool) (string, []string, error) {
	status := getStatus(&model, trainStatus)
	if status == schema.NotStarted || status == schema.Stopped || status == schema.Failed {
		return status, []string{fmt.Sprintf("workflow %v has status %v", model.Name, status)}, nil
	}

	statusPriority := []string{
		schema.Failed, schema.NotStarted, schema.Stopped,
		schema.Starting, schema.InProgress, schema.Complete,
	}

	statuses := map[string][]string{}
	for _, s := range statusPriority {
		statuses[s] = []string{}
	}

	deps, err := listModelDependencies(model.Id, db)
	if err != nil {
		return "", nil, fmt.Errorf("error while getting model status: %w", err)
	}

	for _, dep := range deps {
		status := getStatus(&dep, trainStatus)
		if dep.Id == model.Id {
			statuses[status] = append(statuses[status], fmt.Sprintf("workflow %v has status %v", dep.Id, status))
		} else {
			statuses[status] = append(statuses[status], fmt.Sprintf("the workflow depends on %v which has status %v", dep.Id, status))
		}
	}

	for _, statusType := range statusPriority {
		if len(statuses[statusType]) > 0 {
			return statusType, statuses[statusType], nil
		}
	}

	return "", nil, fmt.Errorf("error finding statuses")
}

func countDownstreamModels(modelId uuid.UUID, db *gorm.DB, activeOnly bool) (int64, error) {
	query := db.Model(&schema.ModelDependency{}).Where("dependency_id = ?", modelId)
	if activeOnly {
		query = query.Joins("Model").Where("deploy_status IN ?", []string{schema.Starting, schema.InProgress, schema.Complete})
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		return 0, schema.NewDbError("counting downstream models", result.Error)
	}

	return count, nil
}

func getJobLogs(db *gorm.DB, modelId uuid.UUID, job string) ([]string, []string, error) {
	deps, err := listModelDependencies(modelId, db)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving job logs: %w", err)
	}

	depIds := make([]uuid.UUID, 0, len(deps))
	for _, dep := range deps {
		depIds = append(depIds, dep.Id)
	}

	var logs []schema.JobLog

	result := db.Where("model_id IN ?", depIds).Where("job = ?", job).Find(&logs)
	if result.Error != nil {
		return nil, nil, schema.NewDbError("retrieving job logs", result.Error)
	}

	errors := make([]string, 0)
	warnings := make([]string, 0)
	for _, log := range logs {
		if log.Level == "error" {
			errors = append(errors, log.Message)
		} else if log.Level == "warning" {
			warnings = append(warnings, log.Message)
		}
	}

	return errors, warnings, nil
}

type StatusResponse struct {
	Status   string   `json:"status"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

func getStatusHandler(w http.ResponseWriter, modelId uuid.UUID, db *gorm.DB, job string) {
	slog.Info("getting status for model", "job", job, "model_id", modelId)

	var res StatusResponse

	err := db.Transaction(func(txn *gorm.DB) error {
		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			return err
		}

		status, _, err := getModelStatus(model, txn, job == "train")
		if err != nil {
			return err
		}
		res.Status = status
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("retrieving model status: %v", err), http.StatusBadRequest)
		return
	}

	errors, warnings, err := getJobLogs(db, modelId, job)
	if err != nil {
		http.Error(w, fmt.Sprintf("retrieving model job messages: %v", err), http.StatusBadRequest)
		return
	}
	res.Errors = errors
	res.Warnings = warnings

	slog.Info("got status for model successfully", "job", job, "model_id", modelId, "status", res.Status)

	utils.WriteJsonResponse(w, res)
}

type updateStatusRequest struct {
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

func updateStatusHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, job string) {
	modelId, err := auth.ModelIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params updateStatusRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if err := schema.CheckValidStatus(params.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("updating status for model", "job", job, "status", params.Status, "model_id", modelId)

	result := db.Model(&schema.Model{Id: modelId}).Update(job+"_status", params.Status)
	if result.Error != nil {
		err := schema.NewDbError("updating model status", result.Error)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(params.Metadata) > 0 {
		metadataJson, err := json.Marshal(params.Metadata)
		if err != nil {
			http.Error(w, fmt.Sprintf("serializing metadata failed: %v", err), http.StatusBadRequest)
			return
		}

		result := db.Save(&schema.ModelAttribute{ModelId: modelId, Key: "metadata", Value: string(metadataJson)})
		if result.Error != nil {
			err := schema.NewDbError("adding model metadata attribute", result.Error)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	slog.Info("updated status for model successfully", "job", job, "status", params.Status, "model_id", modelId)

	utils.WriteSuccess(w)
}

type jobLogRequest struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func jobLogHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, job string) {
	modelId, err := auth.ModelIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params jobLogRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if params.Level != "warning" && params.Level != "error" {
		http.Error(w, fmt.Sprintf("invalid log level '%v', must be 'warning' or 'error'", params.Level), http.StatusBadRequest)
		return
	}

	log := schema.JobLog{Id: uuid.New(), ModelId: modelId, Job: job, Level: params.Level, Message: params.Message}
	result := db.Create(&log)
	if result.Error != nil {
		err := schema.NewDbError("creating job log", result.Error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

func getLogsHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, c nomad.NomadClient, job string) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := schema.GetModel(modelId, db, false, false, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting logs: %v", err), http.StatusBadRequest)
		return
	}

	var jobName string
	if job == "train" {
		jobName = model.TrainJobName()
	} else {
		jobName = model.DeployJobName()
	}

	logs, err := c.JobLogs(jobName)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting logs: %v", err), http.StatusBadRequest)
		return
	}

	utils.WriteJsonResponse(w, logs)
}

// TODO(Anyone): add logic to cleanup configs for failed jobs
func saveConfig(modelId uuid.UUID, jobType string, config interface{}, store storage.Storage) (string, error) {
	trainConfigData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return "", fmt.Errorf("error encoding train config: %w", err)
	}

	configPath := filepath.Join(storage.ModelPath(modelId), fmt.Sprintf("%v_config.json", jobType))
	err = store.Write(configPath, bytes.NewReader(trainConfigData))
	if err != nil {
		return "", fmt.Errorf("error saving %v config: %w", jobType, err)
	}

	// TODO(Any): this is needed because the train/deployment jobs do not use the storage interface
	// in the future once this is standardized it will not be needed
	return filepath.Join(store.Location(), configPath), nil
}

func verifyLicenseForNewJob(nomad nomad.NomadClient, license *licensing.LicenseVerifier, jobCpuUsage int) (string, error) {
	currentCpuUsage, err := nomad.TotalCpuUsage()
	if err != nil {
		return "", err
	}

	licenseData, err := license.Verify(currentCpuUsage + jobCpuUsage)
	if err != nil {
		slog.Error("license verification failed for new job", "error", err)
		return "", err
	}

	return licenseData.BoltLicenseKey, nil
}

func checkForDuplicateModel(db *gorm.DB, modelName string, userId uuid.UUID) error {
	var duplicateModel schema.Model
	result := db.Limit(1).Find(&duplicateModel, "user_id = ? AND name = ?", userId, modelName)
	if result.Error != nil {
		return schema.NewDbError("checking for duplicate model", result.Error)
	}
	if result.RowsAffected != 0 {
		return fmt.Errorf("a model with name %v already exists for user %v", modelName, userId)
	}
	return nil
}

func newModel(modelId uuid.UUID, modelName, modelType string, baseModelId *uuid.UUID, userId uuid.UUID) schema.Model {
	return schema.Model{
		Id:                modelId,
		Name:              modelName,
		Type:              modelType,
		PublishedDate:     time.Now().UTC(),
		TrainStatus:       schema.NotStarted,
		DeployStatus:      schema.NotStarted,
		Access:            schema.Private,
		DefaultPermission: schema.ReadPerm,
		BaseModelId:       baseModelId,
		UserId:            userId,
	}
}

func saveModel(txn *gorm.DB, model schema.Model, user schema.User) error {
	if err := checkForDuplicateModel(txn, model.Name, model.UserId); err != nil {
		return err
	}

	if model.BaseModelId != nil {
		baseModel, err := schema.GetModel(*model.BaseModelId, txn, true, true, false)
		if err != nil {
			return fmt.Errorf("error retrieving specified base model %v: %w", *model.BaseModelId, err)
		}
		if baseModel.Type != model.Type {
			return fmt.Errorf("specified base model has type %v but new model has type %v", baseModel.Type, model.Type)
		}

		perm, err := auth.GetModelPermissions(baseModel.Id, user, txn)
		if err != nil {
			return fmt.Errorf("error verifying permissions for base model %v: %w", baseModel.Id, err)
		}

		if perm < auth.ReadPermission {
			return fmt.Errorf("user %v does not have permission to access base model %v", model.UserId, baseModel.Id)
		}

		if baseModel.TrainStatus != schema.Complete {
			return fmt.Errorf("base model training is not complete, training must be completed before use as base model")
		}

		if model.Attributes == nil && baseModel.Attributes != nil {
			model.Attributes = make([]schema.ModelAttribute, 0, len(baseModel.Attributes))
			for _, attr := range baseModel.Attributes {
				model.Attributes = append(model.Attributes, schema.ModelAttribute{
					ModelId: model.Id,
					Key:     attr.Key,
					Value:   attr.Value,
				})
			}
		}

		if model.Dependencies == nil && baseModel.Dependencies != nil {
			model.Dependencies = make([]schema.ModelDependency, 0, len(baseModel.Dependencies))
			for _, dep := range baseModel.Dependencies {
				model.Dependencies = append(model.Dependencies, schema.ModelDependency{
					ModelId:      model.Id,
					DependencyId: dep.DependencyId,
				})
			}
		}
	}

	result := txn.Create(&model)
	if result.Error != nil {
		return schema.NewDbError("creating model entry", result.Error)
	}

	return nil
}

func checkDiskUsage(storage storage.Storage) error {
	stats, err := storage.Usage()
	if err != nil {
		return err
	}
	oneMib := uint64(1024 * 1024)
	// Either 20% disk needs to be free or 20Gb (in case the disk is very large)
	threshold := min(stats.TotalBytes/5, 20*1024*oneMib)
	if stats.FreeBytes < threshold {
		used := (stats.TotalBytes - stats.FreeBytes) / oneMib
		total := stats.TotalBytes / oneMib
		delta := (threshold - stats.FreeBytes) / oneMib
		return fmt.Errorf("insufficient disk space available, usage: %d/%d Mib, please clear %d Mib", used, total, delta)
	}
	return nil
}

func checkSufficientStorage(storage storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if err := checkDiskUsage(storage); err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInsufficientStorage)
				return
			}
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}

func GenerateApiKey(db *gorm.DB, prefix string) (string, string, error) {
	var secret string
	var secretHash string

	// We are doing a transaction here, to make sure we don't have two keys with same prefix
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := ensurePrefixIsUnique(tx, prefix); err != nil {
			return err
		}

		var err error
		secret, err = generateRandomString(32)
		if err != nil {
			return err
		}

		secretHash = hashSecret(secret)

		return nil
	})
	if err != nil {
		return "", "", err
	}

	fullKey := fmt.Sprintf("%s.%s", prefix, secret)
	return fullKey, secretHash, nil
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	str := base64.RawURLEncoding.EncodeToString(bytes)
	if len(str) < n {
		return "", errors.New("insufficient length in generated string")
	}
	return str[:n], nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func ensurePrefixIsUnique(db *gorm.DB, prefix string) error {
	var count int64
	if err := db.Model(&schema.UserAPIKey{}).
		Where("prefix = ?", prefix).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("prefix collision, please retry")
	}
	return nil
}

func validateApiKey(db *gorm.DB, r *http.Request) (uuid.UUID, error) {

	// Extract the API Key from the header
	fullKey := r.Header.Get("X-API-Key")

	if fullKey == "" {
		return uuid.Nil, nil
	}

	keyParts := strings.SplitN(fullKey, ".", 2)

	if len(keyParts) != 2 {
		return uuid.Nil, nil
	}

	prefix, secret := keyParts[0], keyParts[1]

	// Look up the API Key record in the database
	var record schema.UserAPIKey

	if err := db.Where("prefix = ?", prefix).Preload("Models").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}

	if !record.ExpiryTime.IsZero() {
		if time.Now().After(record.ExpiryTime) {
			return uuid.Nil, nil
		}
	}

	hashed := hashSecret(secret)

	if hashed != record.HashKey {
		return uuid.Nil, nil
	}

	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		return uuid.Nil, err
	}

	for _, model := range record.Models {
		if model.Id == modelId {
			return record.CreatedBy, nil
		}
	}

	return uuid.Nil, nil
}

// ChainMiddlewares chains multiple middlewares into a single middleware
func ChainMiddlewares(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for _, m := range middlewares {
			next = m(next)
		}
		return next
	}
}

func eitherUserOrApiKeyAuthMiddleware(
	db *gorm.DB,
	userAuthMiddlewares chi.Middlewares,
) func(http.Handler) http.Handler {

	userAuthHandler := ChainMiddlewares(userAuthMiddlewares...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")

			if apiKey != "" {

				userID, err := validateApiKey(db, r)

				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				if userID == uuid.Nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				user, err := schema.GetUser(userID, db)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				reqCtx := r.Context()
				reqCtx = context.WithValue(reqCtx, auth.GetUserContextKey(), user)

				next.ServeHTTP(w, r.WithContext(reqCtx))
				return
			}

			userAuthHandler(next).ServeHTTP(w, r)
		})
	}
}
