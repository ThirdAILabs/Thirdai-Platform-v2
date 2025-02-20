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
	"thirdai_platform/model_bazaar/utils"

	"github.com/google/uuid"
)

type NdbTrainRequest struct {
	ModelName             string             `json:"model_name"`
	BaseModelId           *uuid.UUID         `json:"base_model_id"`
	ModelOptions          *config.NdbOptions `json:"model_options"`
	Data                  config.NDBData     `json:"data"`
	JobOptions            config.JobOptions  `json:"job_options"`
	LLMConfig             *config.LLMConfig  `json:"llm_config"`
	GenerativeSupervision bool               `json:"generative_supervision"`
}

func (opts *NdbTrainRequest) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.BaseModelId != nil && opts.ModelOptions != nil {
		allErrors = append(allErrors, fmt.Errorf("only model options or base model can be specified for ndb training"))
	}
	if opts.ModelOptions == nil && opts.BaseModelId == nil {
		opts.ModelOptions = new(config.NdbOptions)
	}

	if opts.ModelOptions != nil {
		allErrors = append(allErrors, opts.ModelOptions.Validate())
	}
	allErrors = append(allErrors, opts.Data.Validate())
	allErrors = append(allErrors, opts.JobOptions.Validate())

	if opts.GenerativeSupervision && opts.LLMConfig == nil {
		allErrors = append(allErrors, fmt.Errorf("llm_config must be specified for generative supervision"))
	}
	if opts.LLMConfig != nil {
		allErrors = append(allErrors, opts.LLMConfig.Validate())
	}

	return errors.Join(allErrors...)
}

func (s *TrainService) TrainNdb(w http.ResponseWriter, r *http.Request) {
	var options NdbTrainRequest
	if !utils.ParseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start ndb training, found the following errors: %v", err), http.StatusUnprocessableEntity)
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.validateUploads(user.Id, options.Data.UnsupervisedFiles); err != nil {
		http.Error(w, fmt.Sprintf("invalid uploads specified: %v", err), GetResponseCode(err))
		return
	}

	if err := s.validateUploads(user.Id, options.Data.SupervisedFiles); err != nil {
		http.Error(w, fmt.Sprintf("invalid uploads specified: %v", err), GetResponseCode(err))
		return
	}

	s.basicTraining(w, r, basicTrainArgs{
		modelName:             options.ModelName,
		modelType:             schema.NdbModel,
		baseModelId:           options.BaseModelId,
		modelOptions:          options.ModelOptions,
		data:                  options.Data,
		jobOptions:            options.JobOptions,
		llmConfig:             options.LLMConfig,
		generativeSupervision: options.GenerativeSupervision,
	})
}

func listLogData[T any](dir string, s storage.Storage) ([]T, error) {
	logFiles, err := s.List(dir)
	if err != nil {
		slog.Error("error listing log files for deployment", "error", err)
		return nil, CodedError(errors.New("error reading log files for retraining"), http.StatusInternalServerError)
	}

	logs := make([]T, 0)
	for _, logFileName := range logFiles {
		if strings.HasSuffix(logFileName, ".jsonl") {
			logFile, err := s.Read(filepath.Join(dir, logFileName))
			if err != nil {
				slog.Error("error reading log file for deployment", "error", err)
				return nil, CodedError(errors.New("error reading log files for retraining"), http.StatusInternalServerError)
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
					slog.Error("error parsing log for deploment", "error", err)
					return nil, CodedError(errors.New("error parsing log files"), http.StatusInternalServerError)
				}
				logs = append(logs, log)
			}
		}
	}

	return logs, nil
}

type ndbInsertionLog struct {
	Documents []config.TrainFile `json:"documents"`
}

type ndbDeletionLog struct {
	DocIds []string `json:"doc_ids"`
}

func (s *TrainService) getNdbRetrainingData(baseModelId uuid.UUID) (config.NDBData, error) {
	deploymentDir := filepath.Join(storage.ModelPath(baseModelId), "deployments/data")

	data := config.NDBData{
		ModelDataType:     config.NdbDataType,
		UnsupervisedFiles: []config.TrainFile{},
		SupervisedFiles: []config.TrainFile{
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

type NdbRetrainRequest struct {
	ModelName   string            `json:"model_name"`
	BaseModelId uuid.UUID         `json:"base_model_id"`
	JobOptions  config.JobOptions `json:"job_options"`
}

func (opts *NdbRetrainRequest) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func (s *TrainService) NdbRetrain(w http.ResponseWriter, r *http.Request) {
	var options NdbRetrainRequest
	if !utils.ParseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start ndb retraining, found the following errors: %v", err), http.StatusUnprocessableEntity)
		return
	}

	slog.Info("starting ndb retraining", "base_model_id", options.BaseModelId, "model_name", options.ModelName)

	data, err := s.getNdbRetrainingData(options.BaseModelId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error collecting retraining data: %v", err), GetResponseCode(err))
		return
	}

	s.basicTraining(w, r, basicTrainArgs{
		modelName:             options.ModelName,
		modelType:             schema.NdbModel,
		baseModelId:           &options.BaseModelId,
		modelOptions:          nil,
		data:                  data,
		jobOptions:            options.JobOptions,
		retraining:            true,
		generativeSupervision: false,
	})

	slog.Info("started ndb retraining succesfully", "base_model_id", options.BaseModelId, "model_name", options.ModelName)
}
