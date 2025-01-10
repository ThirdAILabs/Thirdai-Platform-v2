package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/model_bazaar/utils"
	"time"

	"github.com/google/uuid"
)

type NlpTokenTrainRequest struct {
	ModelName    string                  `json:"model_name"`
	BaseModelId  *uuid.UUID              `json:"base_model_id"`
	ModelOptions *config.NlpTokenOptions `json:"model_options"`
	Data         config.NlpData          `json:"data"`
	TrainOptions config.NlpTrainOptions  `json:"train_options"`
	JobOptions   config.JobOptions       `json:"job_options"`
}

func (opts *NlpTokenTrainRequest) validate() error {
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
	var options NlpTokenTrainRequest
	if !utils.ParseRequestBody(w, r, &options) {
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

type NlpTextTrainRequest struct {
	ModelName    string                 `json:"model_name"`
	BaseModelId  *uuid.UUID             `json:"base_model_id"`
	ModelOptions *config.NlpTextOptions `json:"model_options"`
	Data         config.NlpData         `json:"data"`
	TrainOptions config.NlpTrainOptions `json:"train_options"`
	JobOptions   config.JobOptions      `json:"job_options"`
}

func (opts *NlpTextTrainRequest) validate() error {
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
	var options NlpTextTrainRequest
	if !utils.ParseRequestBody(w, r, &options) {
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

type NlpTrainDatagenRequest struct {
	ModelName   string     `json:"model_name"`
	BaseModelId *uuid.UUID `json:"base_model_id"`

	TaskPrompt  string  `json:"task_prompt"`
	LlmProvider string  `json:"llm_provider"`
	TestSize    float32 `json:"test_size"`

	TokenOptions *config.NlpTokenDatagenOptions `json:"token_options"`
	TextOptions  *config.NlpTextDatagenOptions  `json:"text_options"`

	TrainOptions config.NlpTrainOptions `json:"train_options"`
	JobOptions   config.JobOptions      `json:"job_options"`
}

func (opts *NlpTrainDatagenRequest) modelType() string {
	if opts.TextOptions != nil {
		return schema.NlpTextModel
	}
	return schema.NlpTokenModel
}

func (opts *NlpTrainDatagenRequest) taskOptions() interface{} {
	if opts.TextOptions != nil {
		return opts.TextOptions
	}
	return opts.TokenOptions
}

func (opts *NlpTrainDatagenRequest) modelOptions() interface{} {
	if opts.TextOptions != nil {
		return opts.TextOptions.GetModelOptions()
	}
	return opts.TokenOptions.GetModelOptions()
}

func (opts *NlpTrainDatagenRequest) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.LlmProvider == "" {
		opts.LlmProvider = "openai"
	}

	if opts.TestSize == 0 {
		opts.TestSize = 0.05
	}

	if opts.TokenOptions != nil && opts.TextOptions != nil {
		allErrors = append(allErrors, fmt.Errorf("cannot specify both 'token_options' and 'text_options'"))
	}

	if opts.TokenOptions == nil && opts.TextOptions == nil {
		allErrors = append(allErrors, fmt.Errorf("must specify one of 'token_options' or 'text_options'"))
	}

	if opts.TokenOptions != nil {
		allErrors = append(allErrors, opts.TokenOptions.Validate())
	}

	if opts.TextOptions != nil {
		allErrors = append(allErrors, opts.TextOptions.Validate())
	}

	allErrors = append(allErrors, opts.TrainOptions.Validate())
	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func (s *TrainService) TrainNlpDatagen(w http.ResponseWriter, r *http.Request) {
	var params NlpTrainDatagenRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if err := params.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start nlp-token training, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New()

	slog.Info("starting datagen training", "model_type", params.modelType(), "model_id", modelId, "model_name", params.ModelName)

	license, err := verifyLicenseForNewJob(s.nomad, s.license, params.JobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobToken, err := s.jobAuth.CreateModelJwt(modelId, time.Hour*1000*24)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating job token: %v", err), http.StatusInternalServerError)
		return
	}

	var modelOptions interface{}
	if params.BaseModelId != nil {
		modelOptions = nil
	} else {
		modelOptions = params.modelOptions()
	}

	storageDir, data, err := s.getDatagenData(modelId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      s.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		JobAuthToken:        jobToken,
		ModelId:             modelId,
		ModelType:           params.modelType(),
		BaseModelId:         params.BaseModelId,
		UserId:              user.Id,
		ModelOptions:        modelOptions,
		Data:                data,
		TrainOptions:        params.TrainOptions,
		JobOptions:          params.JobOptions,
		IsRetraining:        false,
	}

	datagenConfig := config.DatagenConfig{
		ModelId:             modelId,
		BaseModelId:         params.BaseModelId,
		ModelBazaarDir:      s.storage.Location(),
		StorageDir:          storageDir,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		TaskPrompt:          params.TaskPrompt,
		LlmProvider:         params.LlmProvider,
		TestSize:            params.TestSize,
		TaskOptions:         params.taskOptions(),
	}

	err = s.createModelAndStartDatagenTraining(params.ModelName, user, trainConfig, datagenConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to start training: %v", err), http.StatusBadRequest)
		return
	}

	slog.Info("started datagen training successfully", "model_type", params.modelType(), "model_id", modelId, "model_name", params.ModelName)

	utils.WriteJsonResponse(w, trainResponse{ModelId: modelId})
}

type NlpTokenRetrainRequest struct {
	ModelName   string    `json:"model_name"`
	BaseModelId uuid.UUID `json:"base_model_id"`

	LlmProvider string  `json:"llm_provider"`
	TestSize    float32 `json:"test_size"`

	TrainOptions config.NlpTrainOptions `json:"train_options"`
	JobOptions   config.JobOptions      `json:"job_options"`
}

func (opts *NlpTokenRetrainRequest) validate() error {
	allErrors := make([]error, 0)

	if opts.ModelName == "" {
		allErrors = append(allErrors, fmt.Errorf("model name must be specified"))
	}

	if opts.LlmProvider == "" {
		opts.LlmProvider = "openai"
	}

	if opts.TestSize == 0 {
		opts.TestSize = 0.05
	}

	allErrors = append(allErrors, opts.TrainOptions.Validate())
	allErrors = append(allErrors, opts.JobOptions.Validate())

	return errors.Join(allErrors...)
}

func (s *TrainService) NlpTokenRetrain(w http.ResponseWriter, r *http.Request) {
	var params NlpTokenRetrainRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if err := params.validate(); err != nil {
		http.Error(w, fmt.Sprintf("unable to start nlp-token training, found the following errors: %v", err), http.StatusBadRequest)
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New()

	slog.Info("starting datagen retraining", "model_type", schema.NlpTokenModel, "model_id", modelId, "model_name", params.ModelName)

	license, err := verifyLicenseForNewJob(s.nomad, s.license, params.JobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobToken, err := s.jobAuth.CreateModelJwt(modelId, time.Hour*1000*24)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating job token: %v", err), http.StatusInternalServerError)
		return
	}

	storageDir, data, err := s.getDatagenData(modelId)
	if err != nil {
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
		BaseModelId:         &params.BaseModelId,
		UserId:              user.Id,
		ModelOptions:        nil,
		Data:                data,
		TrainOptions:        params.TrainOptions,
		JobOptions:          params.JobOptions,
		IsRetraining:        false,
	}

	numSamplesPerTag := 100
	datagenConfig := config.DatagenConfig{
		ModelId:             modelId,
		BaseModelId:         &params.BaseModelId,
		ModelBazaarDir:      s.storage.Location(),
		StorageDir:          storageDir,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		TaskPrompt:          "token_classification",
		LlmProvider:         params.LlmProvider,
		TestSize:            params.TestSize,
		TaskOptions: config.NlpTokenDatagenOptions{
			ModelType:              schema.NlpTokenModel,
			Tags:                   []config.LabelEntity{},
			NumSentencesToGenerate: 1000,
			NumSamplesPerTag:       &numSamplesPerTag,
			Samples:                nil,
			TemplatesPerSample:     10,
			LoadFromStorage:        true,
		},
	}

	err = s.createModelAndStartDatagenTraining(params.ModelName, user, trainConfig, datagenConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to start training: %v", err), http.StatusBadRequest)
		return
	}

	slog.Info("started datagen retraining successfully", "model_type", schema.NlpTokenModel, "model_id", modelId, "model_name", params.ModelName)

	utils.WriteJsonResponse(w, trainResponse{ModelId: modelId})
}

func (s *TrainService) createModelAndStartDatagenTraining(
	modelName string, user schema.User, trainConfig config.TrainConfig, datagenConfig config.DatagenConfig,
) error {
	trainConfigPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
	if err != nil {
		return err
	}

	datagenConfigPath, err := saveConfig(trainConfig.ModelId, "datagen", datagenConfig, s.storage)
	if err != nil {
		return err
	}

	genaiKey, err := s.variables.GenaiKey(datagenConfig.LlmProvider)
	if err != nil {
		return err
	}

	model := createModel(trainConfig.ModelId, modelName, trainConfig.ModelType, trainConfig.BaseModelId, user.Id)

	job := nomad.DatagenTrainJob{
		TrainJob: nomad.TrainJob{
			JobName:    model.TrainJobName(),
			ConfigPath: trainConfigPath,
			Driver:     s.variables.BackendDriver,
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

	return s.saveModelAndStartJob(model, user, job)
}

func (s *TrainService) getDatagenData(modelId uuid.UUID) (string, config.NlpData, error) {
	// TODO(Any): this is needed because the train/deployment jobs do not use the storage interface
	// in the future once this is standardized it will not be needed
	storageDir := filepath.Join(s.storage.Location(), storage.ModelPath(modelId), "generated_data")

	data := config.NlpData{
		SupervisedFiles: []config.FileInfo{
			{Path: filepath.Join(storageDir, "train/train.csv"), Location: "local"},
		},
		TestFiles: []config.FileInfo{
			{Path: filepath.Join(storageDir, "test/test.csv"), Location: "local"},
		},
	}

	err := data.Validate()

	return storageDir, data, err
}
