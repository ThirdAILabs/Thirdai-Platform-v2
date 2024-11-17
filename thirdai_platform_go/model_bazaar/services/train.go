package services

import (
	"fmt"
	"io"
	"log/slog"
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
		r.Post("/ndb-retrain", s.NdbRetrain)
		r.Post("/nlp-token", s.TrainNlpToken)
		r.Post("/nlp-text", s.TrainNlpText)
		r.Post("/upload-data", s.UploadData)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.jobAuth.Verifier())
		r.Use(s.jobAuth.Authenticator())

		r.Post("/update-status", s.UpdateStatus)
		r.Post("/log", s.JobLog)
	})

	r.Route("/{model_id}", func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.ReadPermission))

		r.Get("/status", s.GetStatus)
		r.Get("/logs", s.Logs)
	})

	return r
}

type basicTrainArgs struct {
	modelName    string
	modelType    string
	baseModelId  *string
	modelOptions interface{}
	data         interface{}
	trainOptions interface{}
	jobOptions   config.JobOptions
}

func (s *TrainService) basicTraining(w http.ResponseWriter, r *http.Request, args basicTrainArgs) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelId := uuid.New().String()

	slog.Info("starting training", "model_type", args.modelType, "model_id", modelId, "model_name", args.modelName)

	license, err := verifyLicenseForNewJob(s.nomad, s.license, args.jobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobToken, err := s.jobAuth.CreateToken("model_id", modelId, time.Hour*1000*24)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating job token: %v", err), http.StatusInternalServerError)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      s.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		JobAuthToken:        jobToken,
		ModelId:             modelId,
		ModelType:           args.modelType,
		BaseModelId:         args.baseModelId,
		ModelOptions:        args.modelOptions,
		Data:                args.data,
		TrainOptions:        args.trainOptions,
		JobOptions:          args.jobOptions,
		IsRetraining:        false,
	}

	err = s.createModelAndStartTraining(args.modelName, userId, trainConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting %v training: %v", args.modelType, err), http.StatusBadRequest)
		return
	}

	slog.Info("started training succesfully", "model_type", args.modelType, "model_id", modelId, "model_name", args.modelName)

	writeJsonResponse(w, map[string]string{"model_id": modelId})
}

func (s *TrainService) createModelAndStartTraining(
	modelName, userId string, trainConfig config.TrainConfig,
) error {
	configPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
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

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if model.BaseModelId != nil {
			baseModel, err := schema.GetModel(*model.BaseModelId, txn, false, true, false)
			if err != nil {
				return fmt.Errorf("error retrieving specified base model %v: %w", *model.BaseModelId, err)
			}
			if baseModel.Type != model.Type {
				return fmt.Errorf("specified base model has type %v but new model has type %v", baseModel.Type, model.Type)
			}

			perm, err := auth.GetModelPermissions(baseModel.Id, userId, txn)
			if err != nil {
				return fmt.Errorf("error verifying permissions for base model %v: %w", baseModel.Id, err)
			}

			if perm < auth.ReadPermission {
				return fmt.Errorf("user %v does not have permission to access base model %v", userId, baseModel.Id)
			}
		}

		err = checkForDuplicateModel(txn, model.Name, userId)
		if err != nil {
			slog.Info("unable to start training: duplicate model name", "model_id", model.Id, "model_name", model.Name)
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
			CloudCredentials: s.variables.CloudCredentials,
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

func (s *TrainService) UploadData(w http.ResponseWriter, r *http.Request) {
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

	// TODO(Any): this is needed because the train/deployment jobs do not use the storage interface
	// in the future once this is standardized it will not be needed
	artifactDir = filepath.Join(s.storage.Location(), artifactDir)

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
