package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/model_bazaar/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainService struct {
	db      *gorm.DB
	nomad   nomad.NomadClient
	storage storage.Storage

	userAuth auth.IdentityProvider
	jobAuth  *auth.JwtManager

	license   *licensing.LicenseVerifier
	variables Variables
}

func (s *TrainService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.AuthMiddleware()...)
		r.Use(checkSufficientStorage(s.storage))

		r.Post("/ndb", s.TrainNdb)
		r.Post("/ndb-retrain", s.NdbRetrain)
		r.Post("/nlp-token", s.TrainNlpToken)
		r.Post("/nlp-text", s.TrainNlpText)
		r.Post("/nlp-datagen", s.TrainNlpDatagen)
		r.Post("/nlp-token-retrain", s.NlpTokenRetrain)
		r.Post("/upload-data", s.UploadData)
		r.Post("/verify-doc-dir", s.VerifyDocDir)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.jobAuth.Verifier())
		r.Use(s.jobAuth.Authenticator())

		r.Post("/update-status", s.UpdateStatus)
		r.Post("/log", s.JobLog)
	})

	r.Route("/{model_id}", func(r chi.Router) {
		r.Use(s.userAuth.AuthMiddleware()...)
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
	baseModelId  *uuid.UUID
	modelOptions interface{}
	data         interface{}
	trainOptions interface{}
	jobOptions   config.JobOptions
	retraining   bool
}

type trainResponse struct {
	ModelId uuid.UUID `json:"model_id"`
}

func (s *TrainService) basicTraining(w http.ResponseWriter, r *http.Request, args basicTrainArgs) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model := newModel(uuid.New(), args.modelName, args.modelType, args.baseModelId, user.Id)

	slog.Info("starting training", "model_type", args.modelType, "model_id", model.Id, "model_name", args.modelName)

	license, err := verifyLicenseForNewJob(s.nomad, s.license, args.jobOptions.CpuUsageMhz())
	if err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	jobToken, err := s.jobAuth.CreateModelJwt(model.Id, time.Hour*1000*24)
	if err != nil {
		slog.Error("error creating job token for train job", "error", err)
		http.Error(w, "error setting up train job", http.StatusInternalServerError)
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
		UserId:              user.Id,
		ModelOptions:        args.modelOptions,
		Data:                args.data,
		TrainOptions:        args.trainOptions,
		JobOptions:          args.jobOptions,
		IsRetraining:        args.retraining,
	}

	configPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
	if err != nil {
		slog.Error("error saving train config", "error", err)
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	job := nomad.TrainJob{
		JobName:    model.TrainJobName(),
		ConfigPath: configPath,
		Driver:     s.variables.BackendDriver,
		Resources: nomad.Resources{
			AllocationMhz:       trainConfig.JobOptions.CpuUsageMhz(),
			AllocationMemory:    trainConfig.JobOptions.AllocationMemory,
			AllocationMemoryMax: 60000,
		},
		CloudCredentials: s.variables.CloudCredentials,
	}

	err = s.saveModelAndStartJob(model, user, job)
	if err != nil {
		http.Error(w, fmt.Sprintf("error starting %v training: %v", args.modelType, err), GetResponseCode(err))
		return
	}

	slog.Info("started training succesfully", "model_type", args.modelType, "model_id", model.Id, "model_name", args.modelName)

	utils.WriteJsonResponse(w, trainResponse{ModelId: model.Id})
}

func (s *TrainService) saveModelAndStartJob(model schema.Model, user schema.User, job nomad.Job) error {
	err := s.db.Transaction(func(txn *gorm.DB) error {
		return saveModel(txn, model, user)
	})

	if err != nil {
		return err
	}

	err = s.nomad.StartJob(job)
	if err != nil {
		slog.Error("error starting train job", "error", err)
		return CodedError(errors.New("error starting train job on nomad"), http.StatusInternalServerError)
	}

	result := s.db.Model(&model).Update("train_status", schema.Starting)
	if result.Error != nil {
		slog.Error("sql error updating model train status", "error", err)
		return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
	}

	return nil
}

func getMultipartBoundary(r *http.Request) (string, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return "", CodedError(fmt.Errorf("missing 'Content-Type' header"), http.StatusBadRequest)
	}
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", CodedError(fmt.Errorf("error parsing media type in request: %w", err), http.StatusBadRequest)
	}
	if mediaType != "multipart/form-data" {
		return "", CodedError(fmt.Errorf("expected media type to be 'multipart/form-data'"), http.StatusBadRequest)
	}

	boundary, ok := params["boundary"]
	if !ok {
		return "", CodedError(fmt.Errorf("missing 'boundary' parameter in 'Content-Type' header"), http.StatusBadRequest)
	}

	return boundary, nil
}

func (s *TrainService) UploadData(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	boundary, err := getMultipartBoundary(r)
	if err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	uploadId := uuid.New()

	upload := schema.Upload{
		Id:         uploadId,
		UserId:     user.Id,
		UploadDate: time.Now().UTC(),
	}
	if err := s.db.Create(&upload).Error; err != nil {
		slog.Error("sql error creating upload", "error", err)
		http.Error(w, fmt.Sprintf("unable to create upload: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}

	filenames := make([]string, 0)

	reader := multipart.NewReader(r.Body, boundary)

	saveDir := storage.UploadPath(uploadId)
	if subDir := r.URL.Query().Get("sub_dir"); subDir != "" {
		ok, err := regexp.MatchString(`^\w+$`, subDir)
		if err != nil || !ok {
			http.Error(w, "invalid value for query parameter 'sub_dir', must be alphanumeric or _ characters only", http.StatusBadRequest)
			return
		}
		saveDir = filepath.Join(saveDir, subDir)
	}

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
				http.Error(w, "invalid filename detected in upload files: filename cannot be empty", http.StatusBadRequest)
				return
			}

			filenames = append(filenames, part.FileName())

			newFilepath := filepath.Join(saveDir, part.FileName())
			err := s.storage.Write(newFilepath, part)
			if err != nil {
				slog.Error("error saving uploaded file", "error", err)
				http.Error(w, "error saving uploaded file", http.StatusInternalServerError)
				return
			}
		}
	}

	upload.Files = strings.Join(filenames, ";")
	if err := s.db.Save(&upload).Error; err != nil {
		slog.Error("sql error updating upload file list", "error", err)
		http.Error(w, "error storing upload metadata", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, map[string]uuid.UUID{"upload_id": uploadId})
}

func (s *TrainService) validateUploads(userId uuid.UUID, files []config.TrainFile) error {
	for i, file := range files {
		if file.Location == config.FileLocUpload {
			uploadId, err := uuid.Parse(file.Path)
			if err != nil {
				return CodedError(fmt.Errorf("invalid upload id: %v", file.Path), http.StatusBadRequest)
			}

			var upload schema.Upload
			result := s.db.First(&upload, "id = ?", uploadId)
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return CodedError(fmt.Errorf("upload %v does not exist", uploadId), http.StatusNotFound)
				}
				slog.Error("sql error retrieving upload info", "upload_id", uploadId, "error", result.Error)
				return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
			}

			if upload.UserId != userId {
				return CodedError(fmt.Errorf("user %v does not have permission to access upload %v", userId, uploadId), http.StatusBadRequest)
			}

			files[i].Location = config.FileLocLocal
			files[i].Path = filepath.Join(s.storage.Location(), storage.UploadPath(uploadId))
		}
	}

	return nil
}

func (s *TrainService) GetStatus(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	getStatusHandler(w, modelId, s.db, "train")
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
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := schema.GetModel(modelId, s.db, false, false, false)
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
		slog.Error("error listing train reports", "model_id", modelId, "error", err)
		http.Error(w, "error listing train reports", http.StatusInternalServerError)
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
		slog.Error("no processable train reports found", "model_id", model.Id)
		http.Error(w, fmt.Sprintf("no train reports found for model %v", model.Id), http.StatusBadRequest)
		return
	}

	reportData, err := s.storage.Read(filepath.Join(reportDir, fmt.Sprintf("%d.json", mostRecent)))
	if err != nil {
		slog.Error("error reading train report", "error", err)
		http.Error(w, "error reading train report", http.StatusInternalServerError)
		return
	}
	defer reportData.Close()

	var report interface{}
	err = json.NewDecoder(reportData).Decode(&report)
	if err != nil {
		slog.Error("error parsing train report", "error", err)
		http.Error(w, "error parsing train report", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, report)
}
