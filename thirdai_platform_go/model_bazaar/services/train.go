package services

import (
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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainService struct {
	db      *gorm.DB
	nomad   nomad.NomadClient
	storage storage.Storage

	userAuth *auth.JwtManager
	jobAuth  *auth.JwtManager

	license   *licensing.LicenseVerifier
	variables Variables
}

func (s *TrainService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())

		r.Post("/ndb", s.TrainNdb)
		r.Post("/upload", s.UploadFiles)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.jobAuth.Verifier())
		r.Use(s.jobAuth.Authenticator())

		r.Post("/update-status", s.UpdateStatus)
		r.Post("/log", s.JobLog)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.jobAuth.Verifier())
		r.Use(s.jobAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.ReadPermission))

		r.Get("/status", s.GetStatus)
		r.Get("/logs", s.Logs)
	})

	return r
}

type NdbTrainOptions struct {
	ModelName    string             `json:"model_name"`
	BaseModelId  *string            `json:"base_model_id"`
	ModelOptions *config.NdbOptions `json:"model_options"`
	Data         config.NDBData     `json:"data"`
	JobOptions   config.JobOptions  `json:"job_options"`
}

func (s *TrainService) TrainNdb(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var options NdbTrainOptions
	if !parseRequestBody(w, r, &options) {
		return
	}

	modelId := uuid.New().String()

	if options.ModelOptions == nil && options.BaseModelId == nil {
		http.Error(w, "Either model options or base model must be specified for ndb training", http.StatusBadRequest)
		return
	}
	if options.BaseModelId != nil && options.ModelOptions != nil {
		http.Error(w, "Only model options or base model can be specified for ndb training", http.StatusBadRequest)
		return
	}

	if len(options.Data.SupervisedFiles)+len(options.Data.UnsupervisedFiles) == 0 {
		http.Error(w, "unsupervised or supervised data must be specified for training", http.StatusBadRequest)
		return
	}

	options.JobOptions.EnsureValid()

	license, err := verifyLicenseForNewJob(s.nomad, s.license, options.JobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := s.jobAuth.CreateToken("model_id", modelId, time.Hour*10*24)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating job token: %v", err), http.StatusInternalServerError)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      s.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		UpdateToken:         token,
		ModelId:             modelId,
		BaseModelId:         options.BaseModelId,
		ModelOptions:        options.ModelOptions,
		Data:                options.Data,
		JobOptions:          options.JobOptions,
		IsRetraining:        false,
	}

	err = s.createModelAndStartTraining(options.ModelName, schema.NdbModel, userId, trainConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting ndb training: %v", err), http.StatusBadRequest)
		return
	}

	writeJsonResponse(w, map[string]string{"model_id": modelId})
}

func (s *TrainService) createModelAndStartTraining(
	modelName, modelType, userId string, trainConfig config.TrainConfig,
) error {
	configPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
	if err != nil {
		return err
	}

	model := schema.Model{
		Id:                trainConfig.ModelId,
		Name:              modelName,
		Type:              modelType,
		PublishedDate:     time.Now(),
		TrainStatus:       schema.NotStarted,
		DeployStatus:      schema.NotStarted,
		Access:            schema.Private,
		DefaultPermission: schema.ReadPerm,
		BaseModelId:       trainConfig.BaseModelId,
		UserId:            userId,
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if model.BaseModelId != nil {
			baseModel, err := schema.GetModel(*model.BaseModelId, txn, false, true, false)
			if err != nil {
				return fmt.Errorf("error retrieving specified base model %v: %w", *model.BaseModelId, err)
			}
			if baseModel.Type != model.Type {
				return fmt.Errorf("specified base model has type %v but new model has type %v", baseModel.Type, model.Type)
			}
		}

		err = checkForDuplicateModel(txn, model.Name, userId)
		if err != nil {
			return err
		}

		result := txn.Create(&model)
		if result.Error != nil {
			return schema.NewDbError("creating model entry", result.Error)
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = s.nomad.StartJob(
		nomad.TrainJob{
			JobName:    model.TrainJobName(),
			ConfigPath: configPath,
			Driver:     s.variables.Driver,
			Resources: nomad.Resources{
				AllocationMhz:       trainConfig.JobOptions.CpuUsageMhz(),
				AllocationMemory:    trainConfig.JobOptions.AllocationMemory,
				AllocationMemoryMax: 60000,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error starting train job: %w", err)
	}

	result := s.db.Model(&model).Update("train_status", schema.Starting)
	if result.Error != nil {
		return schema.NewDbError("updating model train status to starting", result.Error)
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

func (s *TrainService) UploadFiles(w http.ResponseWriter, r *http.Request) {
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
			err := s.storage.Write(newFilepath, part)
			if err != nil {
				http.Error(w, fmt.Sprintf("error saving file '%v': %v", part.FileName(), err), http.StatusBadRequest)
				return
			}
		}
	}

	writeJsonResponse(w, map[string]string{"artifact_path": artifactDir})
}

func (s *TrainService) GetStatus(w http.ResponseWriter, r *http.Request) {
	getStatusHandler(w, r, s.db, "train")
}

func (s *TrainService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	updateStatusHandler(w, r, s.db, "train")
}

func (s *TrainService) Logs(w http.ResponseWriter, r *http.Request) {
	getLogsHandler(w, r, s.db, s.nomad, "train")
}

func (s *TrainService) JobLog(w http.ResponseWriter, r *http.Request) {
	jobLogHandler(w, r, s.db, "train")
}
