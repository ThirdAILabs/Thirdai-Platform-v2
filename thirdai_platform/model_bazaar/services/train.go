package services

import (
	"encoding/csv"
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
	"slices"
	"strconv"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrainService struct {
	db                 *gorm.DB
	orchestratorClient orchestrator.Client
	storage            storage.Storage

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
		r.Post("/validate-trainable-csv", s.ValidateTokenTextClassificationCSV)
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
	modelName             string
	modelType             string
	baseModelId           *uuid.UUID
	modelOptions          interface{}
	data                  interface{}
	trainOptions          interface{}
	llmConfig             *config.LLMConfig
	jobOptions            config.JobOptions
	retraining            bool
	generativeSupervision bool
}

type trainResponse struct {
	ModelId uuid.UUID `json:"model_id"`
}

func (s *TrainService) basicTraining(w http.ResponseWriter, r *http.Request, args basicTrainArgs) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model := newModel(uuid.New(), args.modelName, args.modelType, args.baseModelId, user.Id)

	slog.Info("starting training", "model_type", args.modelType, "model_id", model.Id, "model_name", args.modelName)

	license, err := verifyLicenseForNewJob(s.orchestratorClient, s.license, args.jobOptions.CpuUsageMhz())
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
		ModelBazaarDir:        s.storage.Location(),
		LicenseKey:            license,
		ModelBazaarEndpoint:   s.variables.ModelBazaarEndpoint,
		JobAuthToken:          jobToken,
		ModelId:               model.Id,
		ModelType:             args.modelType,
		BaseModelId:           args.baseModelId,
		UserId:                user.Id,
		ModelOptions:          args.modelOptions,
		Data:                  args.data,
		TrainOptions:          args.trainOptions,
		JobOptions:            args.jobOptions,
		IsRetraining:          args.retraining,
		GenerativeSupervision: args.generativeSupervision,
		LLMConfig:             args.llmConfig,
	}

	configPath, err := saveConfig(trainConfig.ModelId, "train", trainConfig, s.storage)
	if err != nil {
		slog.Error("error saving train config", "error", err)
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	job := orchestrator.TrainJob{
		JobName:    model.TrainJobName(),
		ConfigPath: configPath,
		Driver:     s.variables.BackendDriver,
		Resources: orchestrator.Resources{
			AllocationCores:     2,
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

func (s *TrainService) saveModelAndStartJob(model schema.Model, user schema.User, job orchestrator.Job) error {
	err := s.db.Transaction(func(txn *gorm.DB) error {
		return saveModel(txn, model, user)
	})

	if err != nil {
		return err
	}

	err = s.orchestratorClient.StartJob(job)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "invalid value for query parameter 'sub_dir', must be alphanumeric or _ characters only", http.StatusUnprocessableEntity)
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
				http.Error(w, "invalid filename detected in upload files: filename cannot be empty", http.StatusUnprocessableEntity)
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
				return CodedError(fmt.Errorf("user %v does not have permission to access upload %v", userId, uploadId), http.StatusForbidden)
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
	getLogsHandler(w, r, s.db, s.orchestratorClient, "train")
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
		if errors.Is(err, schema.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error retrieving model info: %v", err), http.StatusInternalServerError)
		return
	}

	if model.TrainStatus != schema.Complete {
		http.Error(w, fmt.Sprintf("unable to retrieve train report, model %v has status %v", model.Id, model.TrainStatus), http.StatusUnprocessableEntity)
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
		http.Error(w, fmt.Sprintf("no train reports found for model %v", model.Id), http.StatusUnprocessableEntity)
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
		http.Error(w, fmt.Sprintf("no train reports found for model %v", model.Id), http.StatusUnprocessableEntity)
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

func ValidateCSVHeader(fileHeaders []string, expectedHeaders []string) error {
	if len(fileHeaders) != len(expectedHeaders) {
		return fmt.Errorf("invalid column: expected %v, got %v", expectedHeaders, fileHeaders)
	}

	for _, key := range expectedHeaders {
		if !slices.Contains(fileHeaders, key) {
			return fmt.Errorf("invalid column: expected %v, got %v", expectedHeaders, fileHeaders)
		}
	}
	return nil
}

func (s *TrainService) validateTrainableCSV(filepath string, expectedHeaders []string, targetColumn string, isTokenCSV bool) ([]string, error) {
	file, err := s.storage.Read(filepath)
	if err != nil {
		return nil, CodedError(fmt.Errorf("unable to open file. error: %w", err), http.StatusUnprocessableEntity)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	fileHeaders, err := reader.Read()
	if err != nil {
		return nil, CodedError(fmt.Errorf("unable to read file. error: %w", err), http.StatusUnprocessableEntity)
	}

	// Validate the CSV header
	if err := ValidateCSVHeader(fileHeaders, expectedHeaders); err != nil {
		return nil, CodedError(err, http.StatusUnprocessableEntity)
	}

	targetColIndex := slices.Index(fileHeaders, targetColumn)
	sourceColIndex := 1 - targetColIndex

	labels := make(map[string]bool)

	for {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, CodedError(err, http.StatusUnprocessableEntity)
			}
		}

		if isTokenCSV {
			sourceTokens := strings.Split(line[sourceColIndex], " ")
			targetTokens := strings.Split(line[targetColIndex], " ")
			if len(sourceTokens) != len(targetTokens) {
				return nil, CodedError(fmt.Errorf("number of source tokens: %d ≠ number of target tokens: %d. Invalid line: '%v'", len(sourceTokens), len(targetTokens), strings.Join(line, ",")), http.StatusUnprocessableEntity)
			}
			for _, token := range targetTokens {
				if token != "O" {
					labels[token] = true
				}
			}
		} else {
			labels[line[targetColIndex]] = true
		}
	}

	uniqueLabels := make([]string, 0)
	for key := range labels {
		uniqueLabels = append(uniqueLabels, key)
	}

	return uniqueLabels, nil
}

type TrainableCSVRequest struct {
	UploadId string `json:"upload_id"`
	FileType string `json:"type"`
}

func (s *TrainService) ValidateTokenTextClassificationCSV(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var options TrainableCSVRequest
	if !utils.ParseRequestBody(w, r, &options) {
		return
	}
	switch options.FileType {
	case "text", "token":
	// ok - nothing to do
	default:
		http.Error(w, fmt.Sprintf("%v type is not supported. Supported types: ['text', 'token']", options.FileType), http.StatusUnprocessableEntity)
		return
	}

	// Currently only supporting uploaded file for training text/token classification model. Creating a dummy TrainFile object to validate the upload
	trainConfig := []config.TrainFile{
		{
			Path:     options.UploadId,
			Location: config.FileLocUpload,
			SourceId: nil,
			Options:  nil,
			Metadata: nil,
		},
	}
	if err := s.validateUploads(user.Id, trainConfig); err != nil {
		http.Error(w, fmt.Sprintf("invalid uploads specified: %v", err), GetResponseCode(err))
		return
	}

	UploadID, err := uuid.Parse(options.UploadId)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid upload id: %v", UploadID), GetResponseCode(err))
		return
	}

	fileNames, err := s.storage.List(storage.UploadPath(UploadID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	if len(fileNames) != 1 {
		http.Error(w, fmt.Sprintf("Only one file should be used. found %v files", len(fileNames)), http.StatusUnsupportedMediaType)
		return
	}
	trainableCSVFilePath := filepath.Join(storage.UploadPath(UploadID), fileNames[0])

	if strings.ToLower(filepath.Ext(trainableCSVFilePath)) != ".csv" {
		http.Error(w, "only CSV file is supported", http.StatusUnsupportedMediaType)
		return
	}

	var labels []string
	var validation_err error

	if options.FileType == "text" {
		labels, validation_err = s.validateTrainableCSV(trainableCSVFilePath, []string{"text", "labels"}, "labels", false)
		if validation_err != nil {
			http.Error(w, fmt.Sprintf("Validation failed: %v", validation_err.Error()), GetResponseCode(validation_err))
			return
		}
	} else {
		labels, validation_err = s.validateTrainableCSV(trainableCSVFilePath, []string{"source", "target"}, "target", true)
		if validation_err != nil {
			http.Error(w, fmt.Sprintf("Validation failed: %v", validation_err.Error()), GetResponseCode(validation_err))
			return
		}
	}

	utils.WriteJsonResponse(w, map[string][]string{"labels": labels})
}
