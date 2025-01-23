package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModelService struct {
	db *gorm.DB

	nomad   nomad.NomadClient
	storage storage.Storage

	userAuth          auth.IdentityProvider
	uploadSessionAuth *auth.JwtManager
}

func (s *ModelService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{model_id}", func(r chi.Router) {
		r.Use(s.userAuth.AuthMiddleware()...)

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
	PublishDate  string     `json:"publish_date"`
	UserEmail    string     `json:"user_email"`
	Username     string     `json:"Username"`
	TeamId       *uuid.UUID `json:"team_id"`

	Attributes map[string]string `json:"attributes"`

	Dependencies []ModelDependency `json:"dependencies"`
}

func convertToModelInfo(model schema.Model, db *gorm.DB) (ModelInfo, error) {
	trainStatus, _, err := getModelStatus(model, db, true)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("error retrieving model status: %w", err)
	}
	deployStatus, _, err := getModelStatus(model, db, false)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("error retrieving model status: %w", err)
	}

	attributes := make(map[string]string, len(model.Attributes))
	for _, attr := range model.Attributes {
		attributes[attr.Key] = attr.Value
	}

	deps := make([]ModelDependency, 0, len(model.Dependencies))
	for _, dep := range model.Dependencies {
		deps = append(deps, ModelDependency{
			ModelId:   dep.DependencyId,
			ModelName: dep.Dependency.Name,
			Type:      dep.Dependency.Type,
			Username:  dep.Dependency.User.Username,
		})
	}

	return ModelInfo{
		ModelId:      model.Id,
		ModelName:    model.Name,
		Type:         model.Type,
		Access:       model.Access,
		TrainStatus:  trainStatus,
		DeployStatus: deployStatus,
		PublishDate:  model.PublishedDate.String(),
		UserEmail:    model.User.Email,
		Username:     model.User.Username,
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, err := convertToModelInfo(model, s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.WriteJsonResponse(w, info)
}

func (s *ModelService) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var models []schema.Model
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Preload("Dependencies").Preload("Dependencies.Dependency").Preload("Attributes").Preload("User").Find(&models)
	} else {
		userTeams, err := schema.GetUserTeamIds(user.Id, s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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
		err := schema.NewDbError("listing models", result.Error)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	infos := make([]ModelInfo, 0, len(models))
	for _, model := range models {
		info, err := convertToModelInfo(model, s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		infos = append(infos, info)
	}

	utils.WriteJsonResponse(w, infos)
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
		http.Error(w, fmt.Sprintf("error getting user_id: %v", err), http.StatusBadRequest)
		return
	}

	permission, err := auth.GetModelPermissions(modelId, user, s.db)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving model permissions: %v", err), http.StatusBadRequest)
		return
	}

	expiration, err := s.userAuth.GetTokenExpiration(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving token expiration: %v", err), http.StatusBadRequest)
		return
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
		return 0, schema.NewDbError("counting child models", result.Error)
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
		usedBy, err := countDownstreamModels(modelId, txn, false)
		if err != nil {
			return err
		}
		if usedBy != 0 {
			return fmt.Errorf("cannot delete model %v since it is used as a dependency by %d other models", modelId, usedBy)
		}

		childModels, err := countTrainingChildModels(txn, modelId)
		if err != nil {
			return err
		}
		if childModels != 0 {
			return fmt.Errorf("cannot delete model %v since it is being used as a base model for %d actively training models", modelId, childModels)
		}

		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			return err
		}

		if model.TrainStatus == schema.Starting || model.TrainStatus == schema.InProgress {
			err = s.nomad.StopJob(model.TrainJobName())
			if err != nil {
				return err
			}
		}

		if model.DeployStatus == schema.Starting || model.DeployStatus == schema.InProgress || model.DeployStatus == schema.Complete {
			err = s.nomad.StopJob(model.DeployJobName())
			if err != nil {
				return err
			}
		}

		err = s.storage.Delete(storage.ModelPath(modelId))
		if err != nil {
			return fmt.Errorf("error deleting model date: %v", err)
		}

		err = s.storage.Delete(storage.DataPath(modelId))
		if err != nil {
			return fmt.Errorf("error deleting model date: %v", err)
		}

		// TODO(Nicholas): ensure all relations (deps, attrs, teams, etc) are cleaned up
		result := txn.Delete(&model)
		if result.Error != nil {
			return schema.NewDbError("deleting model", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting model: %v", err), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("error retrieving user id from request: %v", err), http.StatusBadRequest)
		return
	}

	model := newModel(uuid.New(), params.ModelName, schema.UploadInProgress, nil, user.Id)

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkForDuplicateModel(txn, model.Name, model.UserId); err != nil {
			return err
		}

		result := txn.Create(&model)
		if result.Error != nil {
			return schema.NewDbError("creating model entry", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error creating new model: %v", err), http.StatusBadRequest)
		return
	}

	uploadToken, err := s.uploadSessionAuth.CreateModelJwt(model.Id, 10*time.Minute)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating upload token for model %v: %v", model.Name, err), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("error uploading chunk %d to storage: %v", chunkIdx, err), http.StatusBadRequest)
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
		return fmt.Errorf("error listing chunks for model upload: %w", err)
	}

	chunkSet := make(map[string]bool)
	for _, chunk := range chunks {
		chunkSet[chunk] = true
	}

	modelZipfile := filepath.Join(storage.ModelPath(modelId), "model.zip")
	for i := 0; i < len(chunks); i++ {
		chunkPath := strconv.Itoa(i)
		if !chunkSet[chunkPath] {
			return fmt.Errorf("chunk %d is missing", i)
		}

		chunk, err := s.storage.Read(filepath.Join(storage.ModelPath(modelId), "chunks", chunkPath))
		if err != nil {
			return fmt.Errorf("error reading chunk %d: %w", i, err)
		}
		defer chunk.Close()

		err = s.storage.Append(modelZipfile, chunk)
		if err != nil {
			return fmt.Errorf("error writing chunk %d: %w", i, err)
		}
	}

	if err := s.storage.Unzip(modelZipfile); err != nil {
		return fmt.Errorf("error unzipping model, please ensure upload is valid zipfile as returned when downloading a model: %w", err)
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

	return s.db.Transaction(func(txn *gorm.DB) error {
		result := txn.Save(model)
		if result.Error != nil {
			err := schema.NewDbError("updating model on upload commit", result.Error)
			return err
		}

		return nil
	})
}

type ModelMetadata struct {
	Type       string
	Attributes map[string]string
}

func saveModelMetadata(s storage.Storage, model schema.Model) error {
	metadata := ModelMetadata{Type: model.Type, Attributes: model.GetAttributes()}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(metadata); err != nil {
		return fmt.Errorf("error serializing model metadata: %w", err)
	}

	if err := s.Write(storage.ModelMetadataPath(model.Id), buf); err != nil {
		return fmt.Errorf("error saving model metadata: %w", err)
	}

	return nil
}

func (s *ModelService) loadModelMetadata(modelId uuid.UUID) (ModelMetadata, error) {
	rawMetadata, err := s.storage.Read(storage.ModelMetadataPath(modelId))
	if err != nil {
		return ModelMetadata{}, fmt.Errorf("error opening model metadata: %w", err)
	}
	defer rawMetadata.Close()

	var metadata ModelMetadata
	if err := json.NewDecoder(rawMetadata).Decode(&metadata); err != nil {
		return ModelMetadata{}, fmt.Errorf("error parsing model metadata: %w", err)
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
		http.Error(w, fmt.Sprintf("error retrieving model: %v", err), http.StatusBadRequest)
		return
	}

	if err := s.combineChunks(modelId); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO(Anyone): add checksum

	if err := s.completeUpload(&model); err != nil {
		http.Error(w, fmt.Sprintf("error completing model upload: %v", err), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("error retrieving model: %v", err), http.StatusBadRequest)
		return
	}

	if model.TrainStatus != schema.Complete {
		http.Error(w, fmt.Sprintf("can only download model with successfully completed training, model has train status %s", model.TrainStatus), http.StatusBadRequest)
		return
	}

	if len(model.Dependencies) > 0 {
		http.Error(w, "downloading models with dependencies is not yet supported", http.StatusBadRequest)
		return
	}

	if err := saveModelMetadata(s.storage, model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	downloadPath := filepath.Join(storage.ModelPath(model.Id), "model")
	if err := s.storage.Zip(downloadPath); err != nil {
		http.Error(w, fmt.Sprintf("error preparing zipfile for model download: %v", err), http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "http response does not support chunked response.", http.StatusInternalServerError)
		return
	}

	file, err := s.storage.Read(downloadPath + ".zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("error opening model for download: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	buffer := bufio.NewReader(file)
	chunk := make([]byte, 10*1024*1024)

	for {
		readN, err := buffer.Read(chunk)
		isEof := err == io.EOF
		if err != nil && !isEof {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeN, err := w.Write(chunk[:readN])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if writeN != readN {
			http.Error(w, fmt.Sprintf("expected to write %d bytes to stream, wrote %d", readN, writeN), http.StatusInternalServerError)
			return
		}
		flusher.Flush() // Sends chunk

		if isEof {
			break
		}
	}
}

type updateAccessRequest struct {
	Access string `json:"access"`
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := s.db.Model(&schema.Model{Id: modelId}).Update("access", params.Access)
	if result.Error != nil {
		err := schema.NewDbError("updating model access", result.Error)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if result.RowsAffected != 1 {
		http.Error(w, fmt.Sprintf("unable to update access for model %v, please check that the model exists", modelId), http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := s.db.Model(&schema.Model{Id: modelId}).Update("default_permission", params.Permission)
	if result.Error != nil {
		err := schema.NewDbError("updating model default permission", result.Error)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if result.RowsAffected != 1 {
		http.Error(w, fmt.Sprintf("unable to update default permission for model %v, please check that the model exists", modelId), http.StatusBadRequest)
		return
	}

	utils.WriteSuccess(w)
}
