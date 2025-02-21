package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/model_bazaar/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModelService struct {
	db *gorm.DB

	orchestratorClient orchestrator.Client
	storage            storage.Storage

	userAuth          auth.IdentityProvider
	uploadSessionAuth *auth.JwtManager
}

type CreateAPIKeyRequest struct {
	ModelIDs  []uuid.UUID `json:"model_ids"`
	Name      string      `json:"name"`
	Exp       time.Time   `json:"exp"`
	AllModels bool        `json:"all_models"`
}

type APIKeyResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedBy uuid.UUID `json:"created_by"`
	Expiry    time.Time `json:"expiry"`
}

type deleteRequestBody struct {
	APIKeyID uuid.UUID `json:"api_key_id"`
}

func (s *ModelService) Routes() chi.Router {
	r := chi.NewRouter()

	eitherOrMiddleware := eitherUserOrApiKeyAuthMiddleware(s.db, s.userAuth.AuthMiddleware())
	r.Route("/{model_id}", func(r chi.Router) {
		r.Use(eitherOrMiddleware)

		r.Get("/permissions", s.Permissions)

		r.Group(func(r chi.Router) {
			r.Use(auth.ModelPermissionOnly(s.db, auth.ReadPermission))

			r.Get("/", s.Info)
			r.Get("/download", s.Download)
		})

		r.Group(func(r chi.Router) {
			r.Use(auth.ModelPermissionOnly(s.db, auth.OwnerPermission))

			r.Delete("/", s.Delete)
			r.Post("/access", s.UpdateAccess)
			r.Post("/default-permission", s.UpdateDefaultPermission)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.AuthMiddleware()...)

		r.Get("/list", s.List)
		r.Post("/create-api-key", s.CreateAPIKey)
		r.Post("/delete-api-key", s.DeleteAPIKey)
		r.Get("/list-api-keys", s.ListUserAPIKeys)
		r.With(checkSufficientStorage(s.storage)).Post("/upload", s.UploadStart)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.uploadSessionAuth.Verifier())
		r.Use(s.uploadSessionAuth.Authenticator())

		r.Post("/upload/{chunk_idx}", s.UploadChunk)
		r.Post("/upload/commit", s.UploadCommit)
	})

	return r
}

type ModelDependency struct {
	ModelId   uuid.UUID `json:"model_id"`
	ModelName string    `json:"model_name"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
}

type ModelInfo struct {
	ModelId      uuid.UUID  `json:"model_id"`
	ModelName    string     `json:"model_name"`
	Type         string     `json:"type"`
	Access       string     `json:"access"`
	TrainStatus  string     `json:"train_status"`
	DeployStatus string     `json:"deploy_status"`
	PublishDate  time.Time  `json:"publish_date"`
	UserEmail    string     `json:"user_email"`
	Username     string     `json:"username"`
	TeamId       *uuid.UUID `json:"team_id"`

	Attributes map[string]string `json:"attributes"`

	Dependencies []ModelDependency `json:"dependencies"`
}

func convertToModelInfo(model schema.Model, db *gorm.DB) (ModelInfo, error) {
	trainStatus, _, err := getModelStatus(model, db, true)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("error retrieving model train status: %w", err)
	}
	deployStatus, _, err := getModelStatus(model, db, false)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("error retrieving model deploy status: %w", err)
	}

	attributes := make(map[string]string, len(model.Attributes))
	for _, attr := range model.Attributes {
		attributes[attr.Key] = attr.Value
	}

	// Safely handle user information
	var userEmail, username string
	if model.User != nil {
		userEmail = model.User.Email
		username = model.User.Username
	}

	// Safely handle dependencies
	deps := make([]ModelDependency, 0, len(model.Dependencies))
	for _, dep := range model.Dependencies {
		depEntry := ModelDependency{
			ModelId: dep.DependencyId,
		}

		// Check if Dependency exists
		if dep.Dependency != nil {
			depEntry.ModelName = dep.Dependency.Name
			depEntry.Type = dep.Dependency.Type

			// Check if Dependency.User exists
			if dep.Dependency.User != nil {
				depEntry.Username = dep.Dependency.User.Username
			}
		}

		deps = append(deps, depEntry)
	}

	return ModelInfo{
		ModelId:      model.Id,
		ModelName:    model.Name,
		Type:         model.Type,
		Access:       model.Access,
		TrainStatus:  trainStatus,
		DeployStatus: deployStatus,
		PublishDate:  model.PublishedDate,
		UserEmail:    userEmail,
		Username:     username,
		TeamId:       model.TeamId,
		Attributes:   attributes,
		Dependencies: deps,
	}, nil
}

func (s *ModelService) Info(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := schema.GetModel(modelId, s.db, true, true, true)
	if err != nil {
		if errors.Is(err, schema.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info, err := convertToModelInfo(model, s.db)
	if err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	utils.WriteJsonResponse(w, info)
}

func (s *ModelService) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var models []schema.Model
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.
			Preload("Dependencies").
			Preload("Dependencies.Dependency").
			Preload("Dependencies.Dependency.User").
			Preload("Attributes").
			Preload("User").
			Find(&models)
	} else {
		userTeams, err := schema.GetUserTeamIds(user.Id, s.db)
		if err != nil {
			http.Error(w, "error loading user teams to determine model access", http.StatusInternalServerError)
			return
		}
		result = s.db.
			Preload("Dependencies").
			Preload("Dependencies.Dependency").
			Preload("Dependencies.Dependency.User").
			Preload("Attributes").
			Preload("User").
			Where("access = ?", schema.Public).
			Or("access = ? AND user_id = ?", schema.Private, user.Id).
			Or("access = ? AND team_id IN ?", schema.Protected, userTeams).
			Find(&models)
	}

	if result.Error != nil {
		slog.Error("sql error list accessible models", "error", err)
		http.Error(w, fmt.Sprintf("unable to list models: %v", err), http.StatusInternalServerError)
		return
	}

	infos := make([]ModelInfo, 0, len(models))
	for _, model := range models {
		info, err := convertToModelInfo(model, s.db)
		if err != nil {
			http.Error(w, err.Error(), GetResponseCode(err))
			return
		}
		infos = append(infos, info)
	}

	utils.WriteJsonResponse(w, infos)
}

func (s *ModelService) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	req, err := parseCreateAPIKeyRequest(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		parsedModelIDs, err := s.parseAndValidateModelIDs(tx, req.ModelIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return err
		}

		models, err := s.fetchModelsInTransaction(tx, parsedModelIDs, user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		if len(models) != len(parsedModelIDs) {
			http.Error(w, "some model_ids are invalid or do not belong to the user", http.StatusBadRequest)
			return fmt.Errorf("invalid model IDs")
		}

		apiKey, err := s.createAndSaveAPIKeyInTransaction(
			tx,
			req.Name,
			req.Exp,
			user.Id,
			models,
			req.AllModels,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to save API key: %v", err), http.StatusInternalServerError)
			return err
		}

		utils.WriteJsonResponse(w, map[string]string{"api_key": apiKey})
		return nil
	})

	if err != nil {
		// transaction error has already been handled and an appropriate HTTP response sent.
		return
	}
}

func parseCreateAPIKeyRequest(r *http.Request, w http.ResponseWriter) (CreateAPIKeyRequest, error) {
	var req CreateAPIKeyRequest
	if !utils.ParseRequestBody(w, r, &req) {
		return req, errors.New("invalid request body")
	}

	if !req.AllModels && len(req.ModelIDs) == 0 {
		return req, errors.New("model_ids are required if all_models is false")
	}

	if strings.TrimSpace(req.Name) == "" {
		return req, errors.New("name is required")
	}

	if req.Exp.Before(time.Now()) {
		return req, errors.New("api key is already expired")
	}

	return req, nil
}

func (s *ModelService) parseAndValidateModelIDs(tx *gorm.DB, modelIDs []uuid.UUID) ([]uuid.UUID, error) {
	var parsedModelIDs []uuid.UUID
	for _, id := range modelIDs {
		parsedModelIDs = append(parsedModelIDs, id)

		dependencies, err := s.fetchModelDependencies(tx, id)
		if err != nil {
			return nil, fmt.Errorf("unable to get dependencies: %v", err)
		}

		for _, dep := range dependencies {
			parsedModelIDs = append(parsedModelIDs, dep.DependencyId)
		}
	}
	return parsedModelIDs, nil
}

func (s *ModelService) fetchModelDependencies(tx *gorm.DB, modelID uuid.UUID) ([]schema.ModelDependency, error) {
	var dependencies []schema.ModelDependency
	err := tx.Where("model_id = ?", modelID).Find(&dependencies).Error
	if err != nil {
		return nil, err
	}
	return dependencies, nil
}

func (s *ModelService) fetchModelsInTransaction(tx *gorm.DB, modelIDs []uuid.UUID, userID uuid.UUID) ([]schema.Model, error) {
	var models []schema.Model
	err := tx.Preload("Attributes").
		Preload("Dependencies").
		Preload("Dependencies.Dependency").
		Preload("Dependencies.Dependency.User").
		Where("id IN ?", modelIDs).
		Where("user_id = ?", userID).
		Find(&models).Error
	if err != nil {
		slog.Error("sql error fetching models", "error", err)
		return nil, CodedError(schema.ErrModelNotFound, http.StatusInternalServerError)
	}
	return models, nil
}

func (s *ModelService) createAndSaveAPIKeyInTransaction(
	tx *gorm.DB,
	name string,
	expiry time.Time,
	userID uuid.UUID,
	models []schema.Model,
	allModels bool,
) (string, error) {

	apiKey, hashKey, err := generateApiKey()
	if err != nil {
		return "", err
	}

	newAPIKey := schema.UserAPIKey{
		Id:            uuid.New(),
		HashKey:       hashKey,
		Name:          name,
		Models:        models,
		AllModels:     allModels,
		GeneratedTime: time.Now(),
		ExpiryTime:    expiry,
		CreatedBy:     userID,
	}

	if err := tx.Create(&newAPIKey).Error; err != nil {
		slog.Error("sql error creating user api keys", "error", err)
		return "", CodedError(schema.ErrUserAPIKeyNotFound, http.StatusInternalServerError)
	}

	return apiKey, nil
}

func (s *ModelService) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, "invalid or missing user", http.StatusUnauthorized)
		return
	}

	var reqBody deleteRequestBody
	if !utils.ParseRequestBody(w, r, &reqBody) {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if reqBody.APIKeyID == uuid.Nil {
		http.Error(w, "key id is required", http.StatusBadRequest)
		return
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var apiKey schema.UserAPIKey
		if err := tx.First(&apiKey, "id = ?", reqBody.APIKeyID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "API key not found", http.StatusNotFound)
				return err
			}
			http.Error(w, "failed to retrieve API key", http.StatusInternalServerError)
			return err
		}

		if apiKey.CreatedBy != user.Id && !user.IsAdmin {
			http.Error(w, "you do not own this key", http.StatusForbidden)
			return fmt.Errorf("forbidden access")
		}

		if err := tx.Delete(&apiKey).Error; err != nil {
			http.Error(w, "failed to delete API key", http.StatusInternalServerError)
			return err
		}
		return nil
	}); err != nil {
		if err.Error() != "forbidden access" {
			http.Error(w, "an error occurred during the transaction", http.StatusInternalServerError)
		}
		return
	}

	utils.WriteSuccess(w)
}

func (s *ModelService) ListUserAPIKeys(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized: invalid or missing user", http.StatusUnauthorized)
		return
	}

	var apiKeys []APIKeyResponse

	dbQuery := s.db.Model(&schema.UserAPIKey{})

	if !user.IsAdmin {
		dbQuery = dbQuery.Where("created_by = ?", user.Id)
	}

	dbQuery = dbQuery.Select("id, name, created_by, expiry_time as expiry")

	if err := dbQuery.Scan(&apiKeys).Error; err != nil {
		http.Error(w, "failed to retrieve API keys", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, apiKeys)
}

type ModelPermissions struct {
	Read     bool      `json:"read"`
	Write    bool      `json:"write"`
	Owner    bool      `json:"owner"`
	Username string    `json:"username"`
	Exp      time.Time `json:"exp"`
}

func (s *ModelService) Permissions(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting user_id: %v", err), http.StatusInternalServerError)
		return
	}

	permission, err := auth.GetModelPermissions(modelId, user, s.db)
	if err != nil {
		if errors.Is(err, schema.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error retrieving model permissions: %v", err), http.StatusInternalServerError)
		return
	}

	var expiration time.Time

	if expiry, ok := auth.GetAPIKeyExpiry(r.Context()); ok {
		expiration = expiry
	} else {
		expiration, err = s.userAuth.GetTokenExpiration(r)
		if err != nil {
			slog.Error("error retrieving jwt expiration", "error", err)
			http.Error(w, "error retrieving token expiration", http.StatusInternalServerError)
			return
		}
	}

	res := ModelPermissions{
		Read:     permission >= auth.ReadPermission,
		Write:    permission >= auth.WritePermission,
		Owner:    permission >= auth.OwnerPermission,
		Username: user.Username,
		Exp:      expiration,
	}
	utils.WriteJsonResponse(w, res)
}

func countTrainingChildModels(db *gorm.DB, modelId uuid.UUID) (int64, error) {
	var childModels int64
	result := db.Model(&schema.Model{}).
		Where("base_model_id = ?", modelId).
		Where("train_status IN ?", []string{schema.NotStarted, schema.Starting, schema.InProgress}).
		Count(&childModels)

	if result.Error != nil {
		slog.Error("sql error counting child models", "error", result.Error)
		return 0, CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
	}
	return childModels, nil
}

func (s *ModelService) Delete(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		usedBy, err := countDownstreamModels(modelId, txn, false)
		if err != nil {
			return err
		}
		if usedBy != 0 {
			return CodedError(fmt.Errorf("cannot delete model %v since it is used as a dependency by %d other models", modelId, usedBy), http.StatusUnprocessableEntity)
		}

		childModels, err := countTrainingChildModels(txn, modelId)
		if err != nil {
			return err
		}
		if childModels != 0 {
			return CodedError(fmt.Errorf("cannot delete model %v since it is being used as a base model for %d actively training models", modelId, childModels), http.StatusUnprocessableEntity)
		}

		if model.TrainStatus == schema.Starting || model.TrainStatus == schema.InProgress {
			err = s.orchestratorClient.StopJob(model.TrainJobName())
			if err != nil {
				slog.Error("error stopping train job when deleting model", "model_id", modelId, "error", err)
				return CodedError(errors.New("error stopping model train job"), http.StatusInternalServerError)
			}
		}

		if model.DeployStatus == schema.Starting || model.DeployStatus == schema.InProgress || model.DeployStatus == schema.Complete {
			err = s.orchestratorClient.StopJob(model.DeployJobName())
			if err != nil {
				slog.Error("error stopping deploy job when deleting model", "model_id", modelId, "error", err)
				return CodedError(errors.New("error stopping model deploy job"), http.StatusInternalServerError)
			}
		}

		err = s.storage.Delete(storage.ModelPath(modelId))
		if err != nil {
			slog.Error("error deleting model directory", "model_id", modelId, "error", err)
			return CodedError(errors.New("error deleting model data"), http.StatusInternalServerError)
		}

		err = s.storage.Delete(storage.DataPath(modelId))
		if err != nil {
			slog.Error("error deleting model data directory", "model_id", modelId, "error", err)
			return CodedError(errors.New("error deleting model data"), http.StatusInternalServerError)
		}

		result := txn.Delete(&model)
		if result.Error != nil {
			slog.Error("sql error deleting model", "model_id", modelId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting model: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

type UploadStartRequest struct {
	ModelName string `json:"model_name"`
}

func (s *ModelService) UploadStart(w http.ResponseWriter, r *http.Request) {
	var params UploadStartRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving user id from request: %v", err), http.StatusInternalServerError)
		return
	}

	model := newModel(uuid.New(), params.ModelName, schema.UploadInProgress, nil, user.Id)

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkForDuplicateModel(txn, model.Name, model.UserId); err != nil {
			return err
		}

		result := txn.Create(&model)
		if result.Error != nil {
			slog.Error("sql error creating model for upload", "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error creating new model: %v", err), GetResponseCode(err))
		return
	}

	uploadToken, err := s.uploadSessionAuth.CreateModelJwt(model.Id, 10*time.Minute)
	if err != nil {
		slog.Error("error creating upload token", "model_id", model.Id, "error", err)
		http.Error(w, "error creating upload token for model", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, map[string]string{"token": uploadToken})
}

func (s *ModelService) UploadChunk(w http.ResponseWriter, r *http.Request) {
	chunkIdxParam, err := utils.URLParam(r, "chunk_idx")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	chunkIdx, err := strconv.Atoi(chunkIdxParam)
	if err != nil || chunkIdx < 0 {
		http.Error(w, "expected 'chunk_idx' parameter to be an positive integer", http.StatusBadRequest)
		return
	}

	modelId, err := auth.ModelIdFromContext(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving model id from request: %v", err), http.StatusBadRequest)
		return
	}

	path := filepath.Join(storage.ModelPath(modelId), fmt.Sprintf("chunks/%d", chunkIdx))

	err = s.storage.Write(path, r.Body)
	if err != nil {
		slog.Error("error uploading chunk to storage", "model_id", modelId, "chunk_idx", chunkIdx, "error", err)
		http.Error(w, "error uploading chunk to storage", http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}

type uploadCommitResponse struct {
	ModelId   uuid.UUID `json:"model_id"`
	ModelType string    `json:"model_type"`
}

func (s *ModelService) combineChunks(modelId uuid.UUID) error {
	chunks, err := s.storage.List(filepath.Join(storage.ModelPath(modelId), "chunks"))
	if err != nil {
		slog.Error("error listing chunks for model upload", "error", err)
		return CodedError(errors.New("error accessing uploaded data"), http.StatusInternalServerError)
	}

	chunkSet := make(map[string]bool)
	for _, chunk := range chunks {
		chunkSet[chunk] = true
	}

	modelZipfile := filepath.Join(storage.ModelPath(modelId), "model.zip")
	for i := 0; i < len(chunks); i++ {
		chunkPath := strconv.Itoa(i)
		if !chunkSet[chunkPath] {
			return CodedError(fmt.Errorf("chunk %d is missing", i), http.StatusBadRequest)
		}

		chunk, err := s.storage.Read(filepath.Join(storage.ModelPath(modelId), "chunks", chunkPath))
		if err != nil {
			slog.Error("error reading chunk from upload", "model_id", modelId, "chunk_idx", i, "error", err)
			return CodedError(errors.New("error accessing uploaded data"), http.StatusInternalServerError)
		}
		defer chunk.Close()

		err = s.storage.Append(modelZipfile, chunk)
		if err != nil {
			slog.Error("error appending chunk", "model_id", modelId, "chunk_idx", i, "error", err)
			return CodedError(errors.New("error accessing uploaded data"), http.StatusInternalServerError)
		}
	}

	if err := s.storage.Unzip(modelZipfile); err != nil {
		slog.Error("error unzipping model archive", "model_id", modelId, "error", err)
		// This could be because the upload is corrupted, or because an internal error
		return CodedError(errors.New("error opening model archive"), http.StatusInternalServerError)
	}

	return nil
}

func (s *ModelService) completeUpload(model *schema.Model) error {
	metadata, err := s.loadModelMetadata(model.Id)
	if err != nil {
		return err
	}

	model.Type = metadata.Type
	model.TrainStatus = schema.Complete

	if len(metadata.Attributes) > 0 {
		model.Attributes = make([]schema.ModelAttribute, 0, len(metadata.Attributes))
		for key, value := range metadata.Attributes {
			model.Attributes = append(model.Attributes, schema.ModelAttribute{
				ModelId: model.Id,
				Key:     key,
				Value:   value,
			})
		}
	}

	result := s.db.Save(model)
	if result.Error != nil {
		slog.Error("sql error updating model info on upload commit", "model_id", model.Id, "error", result.Error)
		return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
	}

	return nil
}

type ModelMetadata struct {
	Type       string
	Attributes map[string]string
}

func saveModelMetadata(s storage.Storage, model schema.Model) error {
	metadata := ModelMetadata{Type: model.Type, Attributes: model.GetAttributes()}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(metadata); err != nil {
		slog.Error("error serializing metadata for model download", "model_id", model.Id, "error", err)
		return CodedError(errors.New("error creating metadata for download archive"), http.StatusInternalServerError)
	}

	if err := s.Write(storage.ModelMetadataPath(model.Id), buf); err != nil {
		slog.Error("error saving metadata for model download", "model_id", model.Id, "error", err)
		return CodedError(errors.New("error creating metadata for download archive"), http.StatusInternalServerError)
	}

	return nil
}

func (s *ModelService) loadModelMetadata(modelId uuid.UUID) (ModelMetadata, error) {
	rawMetadata, err := s.storage.Read(storage.ModelMetadataPath(modelId))
	if err != nil {
		slog.Error("error opening model metadata", "model_id", modelId, "error", err)
		return ModelMetadata{}, CodedError(errors.New("error loading model metadata"), http.StatusInternalServerError)
	}
	defer rawMetadata.Close()

	var metadata ModelMetadata
	if err := json.NewDecoder(rawMetadata).Decode(&metadata); err != nil {
		slog.Error("error parsing model metadata", "model_id", modelId, "error", err)
		return ModelMetadata{}, CodedError(errors.New("error loading model metadata"), http.StatusInternalServerError)
	}

	return metadata, nil
}

func (s *ModelService) UploadCommit(w http.ResponseWriter, r *http.Request) {
	modelId, err := auth.ModelIdFromContext(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving model id from request: %v", err), http.StatusBadRequest)
		return
	}

	model, err := schema.GetModel(modelId, s.db, false, false, false)
	if err != nil {
		if errors.Is(err, schema.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error retrieving model: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.combineChunks(modelId); err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	// TODO(Anyone): add checksum

	if err := s.completeUpload(&model); err != nil {
		http.Error(w, fmt.Sprintf("error completing model upload: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteJsonResponse(w, uploadCommitResponse{ModelId: model.Id, ModelType: model.Type})
}

func (s *ModelService) Download(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := schema.GetModel(modelId, s.db, true, true, false)
	if err != nil {
		if errors.Is(err, schema.ErrModelNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error retrieving model: %v", err), http.StatusInternalServerError)
		return
	}

	if model.TrainStatus != schema.Complete {
		http.Error(w, fmt.Sprintf("can only download model with successfully completed training, model has train status %s", model.TrainStatus), http.StatusUnprocessableEntity)
		return
	}

	if len(model.Dependencies) > 0 {
		http.Error(w, "downloading models with dependencies is not yet supported", http.StatusUnprocessableEntity)
		return
	}

	if err := saveModelMetadata(s.storage, model); err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	downloadPath := filepath.Join(storage.ModelPath(model.Id), "model")
	if err := s.storage.Zip(downloadPath); err != nil {
		slog.Error("error preparing zipfile for model download", "model_id", modelId, "error", err)
		http.Error(w, "error preparing model download archive", http.StatusInternalServerError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "http response does not support chunked response.", http.StatusInternalServerError)
		return
	}

	file, err := s.storage.Read(downloadPath + ".zip")
	if err != nil {
		slog.Error("error opening model zipfile for download", "model_id", modelId, "error", err)
		http.Error(w, "error reading model download archive", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	buffer := bufio.NewReader(file)
	chunk := make([]byte, 10*1024*1024)

	for {
		readN, err := buffer.Read(chunk)
		isEof := err == io.EOF
		if err != nil && !isEof {
			slog.Error("error reading chunk of model archive", "model_id", modelId, "error", err)
			http.Error(w, "error reading from model download archive", http.StatusInternalServerError)
			return
		}

		writeN, err := w.Write(chunk[:readN])
		if err != nil {
			slog.Error("error writing model download chunk", "model_id", modelId, "error", err)
			http.Error(w, fmt.Sprintf("error sending model download chunk: %v", err), http.StatusInternalServerError)
			return
		}
		if writeN != readN {
			slog.Error("error writing model download chunk", "model_id", modelId, "error", fmt.Sprintf("expected to write %d bytes to stream, wrote %d", readN, writeN))
			http.Error(w, "error sending model download chunk", http.StatusInternalServerError)
			return
		}
		flusher.Flush() // Sends chunk

		if isEof {
			break
		}
	}
}

type updateAccessRequest struct {
	Access string     `json:"access"`
	TeamId *uuid.UUID `json:"team_id"`
}

func (s *ModelService) UpdateAccess(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params updateAccessRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if err := schema.CheckValidAccess(params.Access); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if params.Access == schema.Protected && params.TeamId == nil {
		http.Error(w, "must specifiy team id if changing the model access to protected", http.StatusUnprocessableEntity)
		return
	}

	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving user id from request: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.db.Transaction(func(txn *gorm.DB) error {
		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		model.Access = params.Access
		if params.Access == schema.Protected {
			if err := checkTeamExists(txn, *params.TeamId); err != nil {
				return err
			}

			if !user.IsAdmin {
				if err := checkTeamMember(txn, user.Id, *params.TeamId); err != nil {
					return err
				}
			}

			model.TeamId = params.TeamId
		} else {
			model.TeamId = nil
		}

		if err := txn.Save(model).Error; err != nil {
			slog.Error("sql error updating model access", "model_id", modelId, "error", err)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	}); err != nil {
		slog.Error("error updating model access", "model_id", modelId, "access", params.Access, "team_id", params.TeamId, "error", err)
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

type updateDefaultPermissionRequest struct {
	Permission string `json:"permission"`
}

func (s *ModelService) UpdateDefaultPermission(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params updateDefaultPermissionRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if err := schema.CheckValidPermission(params.Permission); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	result := s.db.Model(&schema.Model{Id: modelId}).Update("default_permission", params.Permission)
	if result.Error != nil {
		slog.Error("sql error updating model default permission", "model_id", modelId, "error", result.Error)
		http.Error(w, fmt.Sprintf("error updating model default permission: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}
	if result.RowsAffected != 1 {
		http.Error(w, schema.ErrModelNotFound.Error(), http.StatusNotFound)
		return
	}

	utils.WriteSuccess(w)
}
