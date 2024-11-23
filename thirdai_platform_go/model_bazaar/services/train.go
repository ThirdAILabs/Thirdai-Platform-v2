package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
		r.Post("/nlp-datagen", s.TrainNlpDatagen)
		r.Post("/nlp-token-retrain", s.NlpTokenRetrain)
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
		r.Get("/report", s.TrainReport)
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
	retraining   bool
}

func (s *TrainService) basicTraining(w http.ResponseWriter, r *http.Request, args basicTrainArgs) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model := createModel(uuid.NewString(), args.modelName, args.modelType, args.baseModelId, userId)

	slog.Info("starting training", "model_type", args.modelType, "model_id", model.Id, "model_name", args.modelName)

	license, err := verifyLicenseForNewJob(s.nomad, s.license, args.jobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jobToken, err := s.jobAuth.CreateToken("model_id", model.Id, time.Hour*1000*24)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating job token: %v", err), http.StatusInternalServerError)
		return
	}

	trainConfig := config.TrainConfig{
		ModelBazaarDir:      s.storage.Location(),
		LicenseKey:          license,
		ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
		JobAuthToken:        jobToken,
		ModelId:             model.Id,
		ModelType:           args.modelType,
		BaseModelId:         args.baseModelId,
		ModelOptions:        args.modelOptions,
		Data:                args.data,
		TrainOptions:        args.trainOptions,
		JobOptions:          args.jobOptions,
		IsRetraining:        args.retraining,
	}

	configPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting training: %v", err), http.StatusBadRequest)
		return
	}

	job := nomad.TrainJob{
		JobName:    model.TrainJobName(),
		ConfigPath: configPath,
		Driver:     s.variables.Driver,
		Resources: nomad.Resources{
			AllocationMhz:       trainConfig.JobOptions.CpuUsageMhz(),
			AllocationMemory:    trainConfig.JobOptions.AllocationMemory,
			AllocationMemoryMax: 60000,
		},
		CloudCredentials: s.variables.CloudCredentials,
	}

	err = s.saveModelAndStartJob(model, job)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting %v training: %v", args.modelType, err), http.StatusBadRequest)
		return
	}

	slog.Info("started training succesfully", "model_type", args.modelType, "model_id", model.Id, "model_name", args.modelName)

	writeJsonResponse(w, map[string]string{"model_id": model.Id})
}

func (s *TrainService) saveModelAndStartJob(model schema.Model, job nomad.Job) error {
	err := s.db.Transaction(func(txn *gorm.DB) error {
		if model.BaseModelId != nil {
			baseModel, err := schema.GetModel(*model.BaseModelId, txn, false, true, false)
			if err != nil {
				return fmt.Errorf("error retrieving specified base model %v: %w", *model.BaseModelId, err)
			}
			if baseModel.Type != model.Type {
				return fmt.Errorf("specified base model has type %v but new model has type %v", baseModel.Type, model.Type)
			}

			perm, err := auth.GetModelPermissions(baseModel.Id, model.UserId, txn)
			if err != nil {
				return fmt.Errorf("error verifying permissions for base model %v: %w", baseModel.Id, err)
			}

			if perm < auth.ReadPermission {
				return fmt.Errorf("user %v does not have permission to access base model %v", model.UserId, baseModel.Id)
			}

			// TODO(Nicholas): Copy dependencies/attributes from base model
		}

		err := checkForDuplicateModel(txn, model.Name, model.UserId)
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

	err = s.nomad.StartJob(job)
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

func (s *TrainService) TrainReport(w http.ResponseWriter, r *http.Request) {
	model, err := schema.GetModel(chi.URLParam(r, "model_id"), s.db, false, false, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving model info: %v", err), http.StatusBadRequest)
		return
	}

	if model.TrainStatus != schema.Complete {
		http.Error(w, fmt.Sprintf("unable to retrieve train report, model %v has status %v", model.Id, model.TrainStatus), http.StatusBadRequest)
		return
	}

	reportDir := filepath.Join(storage.ModelPath(model.Id), "train_reports")

	reports, err := s.storage.List(reportDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to locate train reports: %v", err), http.StatusBadRequest)
		return
	}

	if len(reports) == 0 {
		http.Error(w, fmt.Sprintf("no train reports found for model %v", model.Id), http.StatusBadRequest)
		return
	}

	mostRecent := -1
	for _, report := range reports {
		timestamp, err := strconv.Atoi(strings.TrimSuffix(report, filepath.Ext(report)))
		if err != nil {
			slog.Error("unable to parse train report", "report", report, "error", err)
			continue
		}
		if timestamp > mostRecent {
			mostRecent = timestamp
		}
	}

	if mostRecent <= 0 {
		http.Error(w, fmt.Sprintf("no train reports found for model %v", model.Id), http.StatusBadRequest)
		return
	}

	reportData, err := s.storage.Read(filepath.Join(reportDir, fmt.Sprintf("%d.json", mostRecent)))
	if err != nil {
		http.Error(w, fmt.Sprintf("error reading report file: %v", err), http.StatusBadRequest)
		return
	}
	defer reportData.Close()

	var report interface{}
	err = json.NewDecoder(reportData).Decode(&report)
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing report: %v", err), http.StatusBadRequest)
		return
	}

	writeJsonResponse(w, report)
}
