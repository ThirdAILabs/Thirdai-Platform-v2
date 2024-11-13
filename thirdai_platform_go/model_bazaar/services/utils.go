package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func parseRequestBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dest)
	if err != nil {
		slog.Error("error parsing request body", "error", err)
		http.Error(w, fmt.Sprintf("error parsing request body: %v", err), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("error serializing response body", "error", err)
		http.Error(w, fmt.Sprintf("error serializing response body: %v", err), http.StatusInternalServerError)
	}
}

func writeSuccess(w http.ResponseWriter) {
	writeJsonResponse(w, struct{}{})
}

func listModelDependencies(modelId string, db *gorm.DB) ([]schema.Model, error) {
	visited := map[string]struct{}{}
	models := []schema.Model{}

	var recurse func(string) error

	recurse = func(m string) error {
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

func countDownstreamModels(modelId string, db *gorm.DB, activeOnly bool) (int64, error) {
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

func getJobLogs(db *gorm.DB, modelId, job string) ([]string, []string, error) {
	var logs []schema.JobLog

	result := db.Where("model_id = ?", modelId).Where("job = ?", job).Find(&logs)
	if result.Error != nil {
		return nil, nil, schema.NewDbError("retrieving job messages", result.Error)
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

func getStatusHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, job string) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

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

	writeJsonResponse(w, res)
}

type updateStatusRequest struct {
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

func updateStatusHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, job string) {
	modelId, err := auth.ValueFromContext(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params updateStatusRequest
	if !parseRequestBody(w, r, &params) {
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

	writeSuccess(w)
}

type jobLogRequest struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func jobLogHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, job string) {
	modelId, err := auth.ValueFromContext(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params jobLogRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	log := schema.JobLog{Id: uuid.New().String(), ModelId: modelId, Job: job, Level: params.Level, Message: params.Message}
	result := db.Create(&log)
	if result.Error != nil {
		err := schema.NewDbError("creating job log", result.Error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeSuccess(w)
}

func getLogsHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB, c nomad.NomadClient, job string) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

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

	writeJsonResponse(w, logs)
}

func saveConfig(modelId string, jobType string, config interface{}, store storage.Storage) (string, error) {
	trainConfigData, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("error encoding train config: %w", err)
	}

	configPath := filepath.Join(storage.ModelPath(modelId), fmt.Sprintf("%v_config.json", jobType))
	err = store.Write(configPath, bytes.NewReader(trainConfigData))
	if err != nil {
		return "", fmt.Errorf("error saving %v config: %w", jobType, err)
	}

	return configPath, nil
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

func checkForDuplicateModel(db *gorm.DB, modelName, userId string) error {
	var duplicateModel schema.Model
	result := db.Find(&duplicateModel, "user_id = ? AND name = ?", userId, modelName)
	if result.Error != nil {
		return schema.NewDbError("checking for duplicate model", result.Error)
	}
	if result.RowsAffected != 0 {
		return fmt.Errorf("a model with name %v already exists for user %v", modelName, userId)
	}
	return nil
}
