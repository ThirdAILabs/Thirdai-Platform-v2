package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/model_bazaar/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const thirdaiPlatformKeyPrefix = "thirdai_platform_key"

var (
	ErrMissingAPIKey       = errors.New("API key is missing")
	ErrInvalidAPIKey       = errors.New("API key is invalid")
	ErrExpiredAPIKey       = errors.New("API key has expired")
	ErrAPIKeyModelMismatch = errors.New("API key does not have access to the requested model")
)

type codedError struct {
	err  error
	code int
}

func (e *codedError) Error() string {
	return e.err.Error()
}

func (e *codedError) Unwrap() error {
	return e.err
}

func CodedError(err error, code int) error {
	return &codedError{err: err, code: code}
}

func GetResponseCode(err error) int {
	var cerr *codedError
	if errors.As(err, &cerr) {
		return cerr.code
	}
	slog.Error("non coded error passed to GetResponseCode", "error", err)
	return http.StatusInternalServerError
}

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
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			slog.Error("error listing model dependencies", "error", err)
			return CodedError(fmt.Errorf("error listing model dependencies: %w", err), http.StatusInternalServerError)
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

	return "", nil, CodedError(fmt.Errorf("error finding statuses"), http.StatusInternalServerError)
}

func countDownstreamModels(modelId uuid.UUID, db *gorm.DB, activeOnly bool) (int64, error) {
	query := db.Model(&schema.ModelDependency{}).Where("dependency_id = ?", modelId)
	if activeOnly {
		query = query.Joins("Model").Where("deploy_status IN ?", []string{schema.Starting, schema.InProgress, schema.Complete})
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		slog.Error("sql error counting downstream models", "model_id", modelId, "error", result.Error)
		return 0, CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
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
		slog.Error("sql error listing job logs", "error", result.Error)
		return nil, nil, CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
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
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		status, _, err := getModelStatus(model, txn, job == "train")
		if err != nil {
			return err
		}
		res.Status = status
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("retrieving model status: %v", err), GetResponseCode(err))
		return
	}

	errors, warnings, err := getJobLogs(db, modelId, job)
	if err != nil {
		http.Error(w, fmt.Sprintf("retrieving model job messages: %v", err), GetResponseCode(err))
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
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	slog.Info("updating status for model", "job", job, "status", params.Status, "model_id", modelId)

	err = db.Transaction(func(txn *gorm.DB) error {
		if err := checkModelExists(txn, modelId); err != nil {
			return err
		}

		result := txn.Model(&schema.Model{Id: modelId}).Update(job+"_status", params.Status)
		if result.Error != nil {
			slog.Error("sql error updating model status", "job", job, "status", params.Status, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		if len(params.Metadata) > 0 {
			metadataJson, err := json.Marshal(params.Metadata)
			if err != nil {
				return CodedError(fmt.Errorf("metadata cannot be serialized to json: %w", err), http.StatusBadRequest)
			}

			result := db.Save(&schema.ModelAttribute{ModelId: modelId, Key: "metadata", Value: string(metadataJson)})
			if result.Error != nil {
				slog.Error("sql error adding model metadata attribute", "error", result.Error)
				return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("error updating model status", "job", job, "error", err)
		http.Error(w, err.Error(), GetResponseCode(err))
		return
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
		http.Error(w, fmt.Sprintf("invalid log level '%v', must be 'warning' or 'error'", params.Level), http.StatusUnprocessableEntity)
		return
	}

	log := schema.JobLog{Id: uuid.New(), ModelId: modelId, Job: job, Level: params.Level, Message: params.Message}
	result := db.Create(&log)
	if result.Error != nil {
		slog.Error("sql error creating job log", "error", result.Error)
		http.Error(w, fmt.Sprintf("error creating job log: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

func getLogsHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, c orchestrator.Client, job string) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := schema.GetModel(modelId, db, false, false, false)
	if err != nil {
		if errors.Is(err, schema.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error getting logs: %v", err), http.StatusInternalServerError)
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
		slog.Error("error retrieving job logs from nomad", "error", err)
		http.Error(w, "error getting logs from nomad", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, logs)
}

// TODO(Anyone): add logic to cleanup configs for failed jobs
func saveConfig(modelId uuid.UUID, jobType string, config interface{}, store storage.Storage) (string, error) {
	trainConfigData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		slog.Error("error encoding job config", "error", err)
		return "", CodedError(errors.New("error encoding job config"), http.StatusInternalServerError)
	}

	configPath := filepath.Join(storage.ModelPath(modelId), fmt.Sprintf("%v_config.json", jobType))
	err = store.Write(configPath, bytes.NewReader(trainConfigData))
	if err != nil {
		slog.Error("error saving job config", "error", err)
		return "", CodedError(errors.New("error saving job config"), http.StatusInternalServerError)
	}

	// TODO(Any): this is needed because the train/deployment jobs do not use the storage interface
	// in the future once this is standardized it will not be needed
	return filepath.Join(store.Location(), configPath), nil
}

func verifyLicenseForNewJob(orchestratorClient orchestrator.Client, license *licensing.LicenseVerifier, jobCpuUsage int) (string, error) {
	currentCpuUsage, err := orchestratorClient.TotalCpuUsage()
	if err != nil {
		return "", CodedError(errors.New("unable to get cpu usage from nomad"), http.StatusInternalServerError)
	}

	licenseData, err := license.Verify(currentCpuUsage + jobCpuUsage)
	if err != nil {
		slog.Error("license verification failed for new job", "error", err)
		return "", CodedError(err, http.StatusForbidden)
	}

	return licenseData.BoltLicenseKey, nil
}

func checkForDuplicateModel(db *gorm.DB, modelName string, userId uuid.UUID) error {
	var duplicateModel schema.Model
	result := db.Limit(1).Find(&duplicateModel, "user_id = ? AND name = ?", userId, modelName)
	if result.Error != nil {
		slog.Error("sql error checking for dupliate model", "error", result.Error)
		return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
	}
	if result.RowsAffected != 0 {
		return CodedError(fmt.Errorf("a model with name %v already exists for user %v", modelName, userId), http.StatusConflict)
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
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(fmt.Errorf("error retrieving base model: %w", err), http.StatusInternalServerError)
		}
		if baseModel.Type != model.Type {
			return CodedError(fmt.Errorf("specified base model has type %v but new model has type %v", baseModel.Type, model.Type), http.StatusUnprocessableEntity)
		}

		perm, err := auth.GetModelPermissions(baseModel.Id, user, txn)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(fmt.Errorf("error retrieving permissions for base model %v: %w", baseModel.Id, err), http.StatusInternalServerError)
		}

		if perm < auth.ReadPermission {
			return CodedError(fmt.Errorf("user %v does not have permission to access base model %v", model.UserId, baseModel.Id), http.StatusForbidden)
		}

		if baseModel.TrainStatus != schema.Complete {
			return CodedError(errors.New("base model training is not complete, training must be completed before use as base model"), http.StatusUnprocessableEntity)
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
		slog.Error("sql error creating new model entry", "error", result.Error)
		return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
	}

	return nil
}

func checkDiskUsage(storage storage.Storage) error {
	stats, err := storage.Usage()
	if err != nil {
		slog.Error("unable to get disk usage from storage", "error", err)
		return CodedError(errors.New("unable to get disk usage"), http.StatusInternalServerError)
	}
	oneMib := uint64(1024 * 1024)
	// Either 20% disk needs to be free or 20Gb (in case the disk is very large)
	threshold := min(stats.TotalBytes/5, 20*1024*oneMib)
	if stats.FreeBytes < threshold {
		used := (stats.TotalBytes - stats.FreeBytes) / oneMib
		total := stats.TotalBytes / oneMib
		delta := (threshold - stats.FreeBytes) / oneMib
		return CodedError(fmt.Errorf("insufficient disk space available, usage: %d/%d Mib, please clear %d Mib", used, total, delta), http.StatusInsufficientStorage)
	}
	return nil
}

func checkSufficientStorage(storage storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if err := checkDiskUsage(storage); err != nil {
				slog.Error(err.Error())
				http.Error(w, err.Error(), GetResponseCode(err))
				return
			}
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handler)
	}
}

func checkTeamExists(txn *gorm.DB, teamId uuid.UUID) error {
	if _, err := schema.GetTeam(teamId, txn); err != nil {
		if errors.Is(err, schema.ErrTeamNotFound) {
			return CodedError(err, http.StatusNotFound)
		}
		return CodedError(err, http.StatusInternalServerError)
	}
	return nil
}

func checkUserExists(txn *gorm.DB, userId uuid.UUID) error {
	if _, err := schema.GetUser(userId, txn); err != nil {
		if errors.Is(err, schema.ErrUserNotFound) {
			return CodedError(err, http.StatusNotFound)
		}
		return CodedError(err, http.StatusInternalServerError)
	}
	return nil
}

func checkModelExists(txn *gorm.DB, modelId uuid.UUID) error {
	if _, err := schema.GetModel(modelId, txn, false, false, false); err != nil {
		if errors.Is(err, schema.ErrModelNotFound) {
			return CodedError(err, http.StatusNotFound)
		}
		return CodedError(err, http.StatusInternalServerError)
	}
	return nil
}

func checkTeamMember(txn *gorm.DB, userId, teamId uuid.UUID) error {
	if _, err := schema.GetUserTeam(teamId, userId, txn); err != nil {
		if errors.Is(err, schema.ErrUserTeamNotFound) {
			return CodedError(errors.New("user is not a member of team"), http.StatusNotFound)
		}
		return CodedError(err, http.StatusInternalServerError)
	}
	return nil
}
