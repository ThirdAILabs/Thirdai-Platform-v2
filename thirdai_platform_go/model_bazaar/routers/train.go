package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainRouter struct {
	db      *gorm.DB
	nomad   nomad.NomadClient
	storage storage.Storage
	license licensing.LicenseVerifier
}

type ndbTrainOptions struct {
	ModelName    string             `json:"model_name"`
	BaseModelId  *string            `json:"base_model_id"`
	ModelOptions *config.NdbOptions `json:"model_options"`
	Data         config.NDBData     `json:"data"`
	JobOptions   config.JobOptions  `json:"job_options"`
}

func (t *TrainRouter) TrainNdb(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var options ndbTrainOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	modelId := uuid.New().String()

	if options.ModelOptions == nil {
		// TODO(Nicholas): set default options
	}

	if len(options.Data.SupervisedFiles)+len(options.Data.UnsupervisedFiles) == 0 {
		http.Error(w, "unsupervised or supervised data must be specified for training", http.StatusBadRequest)
		return
	}

	options.JobOptions.EnsureValid()

	// TODO(Nicholas): Get total cpu usage
	license, err := t.verifyLicenseForTraining(options.JobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      t.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: "TODO(Nicholas)",
		ModelId:             modelId,
		BaseModelId:         options.BaseModelId,
		ModelOptions:        options.ModelOptions,
		Data:                options.Data,
		JobOptions:          options.JobOptions,
		IsRetraining:        false,
	}

	subtype, err := options.ModelOptions.SubType()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = t.createModelAndStartTraining(options.ModelName, schema.NdbType, subtype, userId, trainConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJsonResponse(w, map[string]string{"model_id": modelId})
}

func (t *TrainRouter) verifyLicenseForTraining(jobCpuUsage int) (string, error) {
	currentCpuUsage, err := t.nomad.TotalCpuUsage()
	if err != nil {
		return "", err
	}

	license, err := t.license.Verify(currentCpuUsage + jobCpuUsage)
	if err != nil {
		return "", err
	}

	return license.BoltLicenseKey, nil
}

func (t *TrainRouter) createModelAndStartTraining(
	modelName, modelType, modelSubtype, userId string, trainConfig config.TrainConfig,
) error {
	trainConfigData, err := json.Marshal(trainConfig)
	if err != nil {
		return fmt.Errorf("error encoding train config: %w", err)
	}

	configPath := filepath.Join(storage.ModelPath(trainConfig.ModelId), "train_config.json")
	err = t.storage.Write(configPath, bytes.NewReader(trainConfigData))
	if err != nil {
		return fmt.Errorf("error saving train config: %w", err)
	}

	model := schema.Model{
		Id:                trainConfig.ModelId,
		Name:              modelName,
		Type:              modelType,
		Subtype:           modelSubtype,
		PublishedDate:     time.Now(),
		TrainStatus:       schema.NotStarted,
		DeployStatus:      schema.NotStarted,
		Access:            schema.Private,
		DefaultPermission: schema.ReadPerm,
		BaseModelId:       trainConfig.BaseModelId,
		UserId:            userId,
	}

	err = t.db.Transaction(func(db *gorm.DB) error {
		if model.BaseModelId != nil {
			exists, err := schema.ModelExists(db, *model.BaseModelId)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("base model %v does not exist", *model.BaseModelId)
			}
		}

		var duplicateModel schema.Model
		result := db.Find(&duplicateModel, "user_id = ? AND name = ?", userId, model.Name)
		if result.Error != nil {
			return fmt.Errorf("database error: %w", result.Error)
		}
		if result.RowsAffected != 0 {
			return fmt.Errorf("a model with name %v already exists", model.Name)
		}

		result = db.Create(&model)
		if result.Error != nil {
			return fmt.Errorf("database error: %v", result.Error)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating entry for new model: %w", err)
	}

	err = t.nomad.StartJob(
		"train_job.hcl", // TODO(Nicholas): template dir path?
		nomad.TrainJob{
			JobName:      nomad.TrainJobName(model),
			TrainScript:  "TODO(Nicholas).py",
			ConfigPath:   configPath,
			PythonPath:   "TODO(Nicholas)",
			PlatformType: "TODO(Nicholas)",
			Platform:     nomad.DockerPlatform{}, // TODO(Nicholas)
			Resources: nomad.Resource{
				AllocationMhz:       trainConfig.JobOptions.CpuUsageMhz(),
				AllocationMemory:    trainConfig.JobOptions.AllocationMemory,
				AllocationMemoryMax: 60000,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to start training job: %w", err)
	}

	model.TrainStatus = schema.Starting

	// TODO(Nicholas): prevent users from deleting base model until training is complete
	result := t.db.Save(&model)
	if result.Error != nil {
		return fmt.Errorf("database error: %w", result.Error)
	}

	return nil
}

func getMultipartBoundary(r *http.Request) (string, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return "", fmt.Errorf("missing 'Content-Type' header")
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("error parsing media type in request: %w", err)
	}
	if mediaType != "multipart/form-data" {
		return "", fmt.Errorf("expected media type to be 'multipart/form-data'")
	}

	boundary, ok := params["boundary"]
	if !ok {
		return "", fmt.Errorf("missing 'boundary' parameter in 'Content-Type' header")
	}

	return boundary, nil
}

func (t *TrainRouter) UploadFiles(w http.ResponseWriter, r *http.Request) {
	boundary, err := getMultipartBoundary(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reader := multipart.NewReader(r.Body, boundary)

	artifactDir := storage.DataPath(uuid.New().String())

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing multipart request: %v", err), http.StatusBadRequest)
			return
		}
		defer part.Close()

		if part.FormName() == "files" {
			if part.FileName() == "" {
				http.Error(w, fmt.Sprintf("invalid filename detected in upload files"), http.StatusBadRequest)
				return
			}
			newFilepath := filepath.Join(artifactDir, part.FileName())
			err := t.storage.Write(newFilepath, part)
			if err != nil {
				http.Error(w, fmt.Sprintf("error saving file '%v': %v", part.FileName(), err), http.StatusBadRequest)
				return
			}
		}
	}

	writeJsonResponse(w, map[string]string{"artifact_path": artifactDir})
}
