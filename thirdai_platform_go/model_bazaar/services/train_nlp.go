package services

import (
	"errors"
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/schema"
)

type NlpTokenTrainOptions struct {
	ModelName    string                  `json:"model_name"`
	BaseModelId  *string                 `json:"base_model_id"`
	ModelOptions *config.NlpTokenOptions `json:"model_options"`
	Data         config.NlpData          `json:"data"`
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
		jobOptions:   options.JobOptions,
	})
}

type NlpTextTrainOptions struct {
	ModelName    string                 `json:"model_name"`
	BaseModelId  *string                `json:"base_model_id"`
	ModelOptions *config.NlpTextOptions `json:"model_options"`
	Data         config.NlpData         `json:"data"`
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
		jobOptions:   options.JobOptions,
	})
}
