package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/google/uuid"
)

type NlpTokenTrainOptions struct {
	ModelName    string                  `json:"model_name"`
	BaseModelId  *string                 `json:"base_model_id"`
	ModelOptions *config.NlpTokenOptions `json:"model_options"`
	Data         config.NlpData          `json:"data"`
	TrainOptions config.NlpTrainOptions  `json:"train_options"`
	JobOptions   config.JobOptions       `json:"job_options"`
}

func (opts *NlpTokenTrainOptions) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.BaseModelId != nil && opts.ModelOptions != nil {
		allErrors = append(allErrors, fmt.Errorf("Only model options or base model can be specified for training"))
	}
	if opts.ModelOptions == nil && opts.BaseModelId == nil {
		opts.ModelOptions = new(config.NlpTokenOptions)
	}

	if opts.ModelOptions != nil {
		allErrors = append(allErrors, opts.ModelOptions.Validate())
	}

	allErrors = append(allErrors, opts.Data.Validate())
	allErrors = append(allErrors, opts.TrainOptions.Validate())
	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func (s *TrainService) TrainNlpToken(w http.ResponseWriter, r *http.Request) {
	var options NlpTokenTrainOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start nlp-token training, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	s.basicTraining(w, r, basicTrainArgs{
		modelName:    options.ModelName,
		modelType:    schema.NlpTokenModel,
		baseModelId:  options.BaseModelId,
		modelOptions: options.ModelOptions,
		data:         options.Data,
		trainOptions: options.TrainOptions,
		jobOptions:   options.JobOptions,
	})
}

type NlpTextTrainOptions struct {
	ModelName    string                 `json:"model_name"`
	BaseModelId  *string                `json:"base_model_id"`
	ModelOptions *config.NlpTextOptions `json:"model_options"`
	Data         config.NlpData         `json:"data"`
	TrainOptions config.NlpTrainOptions `json:"train_options"`
	JobOptions   config.JobOptions      `json:"job_options"`
}

func (opts *NlpTextTrainOptions) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.BaseModelId != nil && opts.ModelOptions != nil {
		allErrors = append(allErrors, fmt.Errorf("Only model options or base model can be specified for training"))
	}
	if opts.ModelOptions == nil && opts.BaseModelId == nil {
		opts.ModelOptions = new(config.NlpTextOptions)
	}

	if opts.ModelOptions != nil {
		allErrors = append(allErrors, opts.ModelOptions.Validate())
	}

	allErrors = append(allErrors, opts.Data.Validate())
	allErrors = append(allErrors, opts.TrainOptions.Validate())
	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func (s *TrainService) TrainNlpText(w http.ResponseWriter, r *http.Request) {
	var options NlpTextTrainOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start nlp-text training, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	s.basicTraining(w, r, basicTrainArgs{
		modelName:    options.ModelName,
		modelType:    schema.NlpTextModel,
		baseModelId:  options.BaseModelId,
		modelOptions: options.ModelOptions,
		data:         options.Data,
		trainOptions: options.TrainOptions,
		jobOptions:   options.JobOptions,
	})
}

type NlpTokenTrainDatagenOptions struct {
	ModelName      string                 `json:"model_name"`
	BaseModelId    *string                `json:"base_model_id"`
	DatagenOptions config.DatagenConfig   `json:"datagen_options"`
	TrainOptions   config.NlpTrainOptions `json:"train_options"`
	JobOptions     config.JobOptions      `json:"job_options"`
}

func (opts *NlpTokenTrainDatagenOptions) validate() error {
	return nil
}

func (s *TrainService) TrainNlpTokenDatagen(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var options NlpTokenTrainDatagenOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	if err := options.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start nlp-token dategen training, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	slog.Info("starting training", "model_type", schema.NlpTokenModel, "model_id", modelId, "model_name", options.ModelName)

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

	var modelOptions *config.NlpTokenOptions
	if options.BaseModelId != nil {
		modelOptions = nil
	} else {
		modelOptions = &config.NlpTokenOptions{
			ModelType:    schema.NlpTokenModel,
			TargetLabels: []string{}, // todo use the data gen labels
			SourceColumn: "source",
			TargetColumn: "target",
			DefaultTag:   "O",
		}
	}

	data := config.NlpData{
		SupervisedFiles: []config.FileInfo{
			// todo use the data gen files
		},
	}

	if err := data.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      s.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		JobAuthToken:        jobToken,
		ModelId:             modelId,
		ModelType:           schema.NlpTokenModel,
		BaseModelId:         options.BaseModelId,
		ModelOptions:        modelOptions,
		Data:                data,
		TrainOptions:        options.TrainOptions,
		JobOptions:          options.JobOptions,
		IsRetraining:        false,
	}

	err = s.createModelAndStartDatagenTraining(options.ModelName, userId, trainConfig, options.DatagenOptions, "")
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to start training: %v", err), http.StatusBadRequest)
		return
	}

	writeJsonResponse(w, map[string]string{"model_id": modelId})
}

func (s *TrainService) createModelAndStartDatagenTraining(
	modelName, userId string, trainConfig config.TrainConfig, datagenConfig config.DatagenConfig, genaiKey string,
) error {
	trainConfigPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
	if err != nil {
		return err
	}

	datagenConfigPath, err := saveConfig(trainConfig.ModelId, "datagen", datagenConfig, s.storage)
	if err != nil {
		return err
	}

	model := schema.Model{
		Id:                trainConfig.ModelId,
		Name:              modelName,
		Type:              trainConfig.ModelType,
		PublishedDate:     time.Now(),
		TrainStatus:       schema.NotStarted,
		DeployStatus:      schema.NotStarted,
		Access:            schema.Private,
		DefaultPermission: schema.ReadPerm,
		BaseModelId:       trainConfig.BaseModelId,
		UserId:            userId,
	}

	job := nomad.DatagenTrainJob{
		TrainJob: nomad.TrainJob{
			JobName:    model.TrainJobName(),
			ConfigPath: trainConfigPath,
			Driver:     s.variables.Driver,
			Resources: nomad.Resources{
				AllocationMhz:       trainConfig.JobOptions.CpuUsageMhz(),
				AllocationMemory:    trainConfig.JobOptions.AllocationMemory,
				AllocationMemoryMax: 60000,
			},
			CloudCredentials: s.variables.CloudCredentials,
		},
		DatagenConfigPath: datagenConfigPath,
		GenaiKey:          genaiKey,
	}

	return s.saveModelAndStartJob(model, job)
}
