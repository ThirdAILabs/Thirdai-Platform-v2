package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/google/uuid"
)

type NdbTrainOptions struct {
	ModelName    string             `json:"model_name"`
	BaseModelId  *string            `json:"base_model_id"`
	ModelOptions *config.NdbOptions `json:"model_options"`
	Data         config.NDBData     `json:"data"`
	JobOptions   config.JobOptions  `json:"job_options"`
}

func (opts *NdbTrainOptions) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.BaseModelId != nil && opts.ModelOptions != nil {
		allErrors = append(allErrors, fmt.Errorf("Only model options or base model can be specified for ndb training"))
	}
	if opts.ModelOptions == nil && opts.BaseModelId == nil {
		opts.ModelOptions = new(config.NdbOptions)
	}

	if opts.ModelOptions != nil {
		allErrors = append(allErrors, opts.ModelOptions.Validate())
	}
	allErrors = append(allErrors, opts.Data.Validate())
	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func (s *TrainService) TrainNdb(w http.ResponseWriter, r *http.Request) {
	var options NdbTrainOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start ndb training, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	s.basicTraining(w, r, basicTrainArgs{
		modelName:    options.ModelName,
		modelType:    schema.NdbModel,
		baseModelId:  options.BaseModelId,
		modelOptions: options.ModelOptions,
		data:         options.Data,
		jobOptions:   options.JobOptions,
	})
}

type NdbRetrainOptions struct {
	ModelName   string            `json:"model_name"`
	BaseModelId string            `json:"base_model_id"`
	JobOptions  config.JobOptions `json:"job_options"`
}

func (opts *NdbRetrainOptions) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.BaseModelId == "" {
		allErrors = append(allErrors, fmt.Errorf("base model id must be specified"))
	}

	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func listLogData[T any](dir string, s storage.Storage) ([]T, error) {
	logFiles, err := s.List(dir)
	if err != nil {
		return nil, fmt.Errorf("error listing log files for deployment: %w", err)
	}

	logs := make([]T, 0)
	for _, logFileName := range logFiles {
		if strings.HasSuffix(logFileName, ".jsonl") {
			logFile, err := s.Read(filepath.Join(dir, logFileName))
			if err != nil {
				return nil, fmt.Errorf("error reading log file from deployment: %w", err)
			}
			defer logFile.Close()

			decoder := json.NewDecoder(logFile)
			for {
				var log T
				err := decoder.Decode(&log)
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("error parsing logs from deployment: %w", err)
				}
				logs = append(logs, log)
			}
		}
	}

	return logs, nil
}

type ndbInsertionLog struct {
	Documents []config.FileInfo `json:"documents"`
}

type ndbDeletionLog struct {
	DocIds []string `json:"doc_ids"`
}

func (s *TrainService) getNdbRetrainingData(baseModelId string) (config.NDBData, error) {
	deploymentDir := filepath.Join(storage.ModelPath(baseModelId), "deployments/data")

	data := config.NDBData{
		UnsupervisedFiles: []config.FileInfo{},
		SupervisedFiles: []config.FileInfo{
			// TODO(Any): this is needed because the train/deployment jobs do not use the storage interface
			// in the future once this is standardized it will not be needed
			{Path: filepath.Join(s.storage.Location(), deploymentDir, "feedback"), Location: config.FileLocLocal, Options: map[string]interface{}{}},
		},
		Deletions: []string{},
	}

	insertionLogs, err := listLogData[ndbInsertionLog](filepath.Join(deploymentDir, "insertions"), s.storage)
	if err != nil {
		return config.NDBData{}, err
	}
	for _, insertLog := range insertionLogs {
		data.UnsupervisedFiles = append(data.UnsupervisedFiles, insertLog.Documents...)
	}

	deletionLogs, err := listLogData[ndbDeletionLog](filepath.Join(deploymentDir, "deletions"), s.storage)
	if err != nil {
		return config.NDBData{}, err
	}
	for _, deleteLog := range deletionLogs {
		data.Deletions = append(data.Deletions, deleteLog.DocIds...)
	}

	return data, nil
}

func (s *TrainService) NdbRetrain(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var options NdbRetrainOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start ndb retraining, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	slog.Info("starting ndb retraining", "model_id", modelId, "model_name", options.ModelName)

	license, err := verifyLicenseForNewJob(s.nomad, s.license, options.JobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobToken, err := s.jobAuth.CreateToken("model_id", modelId, time.Hour*1000*24)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating job token: %v", err), http.StatusInternalServerError)
		return
	}

	data, err := s.getNdbRetrainingData(options.BaseModelId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error collecting retraining data: %v", err), http.StatusBadRequest)
	}

	if err := data.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("data validation failed for ndb retrainin: %v", err), http.StatusBadRequest)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      s.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		JobAuthToken:        jobToken,
		ModelId:             modelId,
		ModelType:           schema.NdbModel,
		BaseModelId:         &options.BaseModelId,
		ModelOptions:        nil,
		Data:                data,
		JobOptions:          options.JobOptions,
		IsRetraining:        true,
	}

	err = s.createModelAndStartTraining(options.ModelName, userId, trainConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting ndb training: %v", err), http.StatusBadRequest)
		return
	}

	slog.Info("started ndb training succesfully", "model_id", modelId, "model_name", options.ModelName)

	writeJsonResponse(w, map[string]string{"model_id": modelId})
}
