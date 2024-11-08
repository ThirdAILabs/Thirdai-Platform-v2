package services

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type ModelService struct {
	db *gorm.DB

	nomad   nomad.NomadClient
	storage storage.Storage

	userAuth    *auth.JwtManager
	sessionAuth *auth.JwtManager
}

func (s *ModelService) Routers() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.ReadPermission))

		r.Get("/details", s.Details)
		r.Get("/download", s.Download)

	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.OwnerPermission))

		r.Post("/delete", s.Delete)
		r.Post("/update-access", s.UpdateAccess)
		r.Post("/update-default-permission", s.UpdateDefaultPermission)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())

		r.Get("/list", s.List)
		r.Post("/save-deployed", s.SaveDeployed)
		r.Post("/upload-token", s.UploadToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.sessionAuth.Verifier())
		r.Use(s.sessionAuth.Authenticator())

		r.Post("/upload-chunk", s.UploadChunk)
		r.Post("/upload-commit", s.UploadCommit)
	})

	return r
}

type ModelDependency struct {
	ModelId   string `json:"model_id"`
	ModelName string `json:"model_name"`
	Type      string `json:"type"`
	Username  string `json:"username"`
}

type ModelInfo struct {
	ModelId      string  `json:"model_id"`
	ModelName    string  `json:"model_name"`
	Type         string  `json:"type"`
	Access       string  `json:"access"`
	TrainStatus  string  `json:"train_status"`
	DeployStatus string  `json:"deploy_status"`
	PublishDate  string  `json:"publish_date"`
	UserEmail    string  `json:"user_email"`
	Username     string  `json:"Username"`
	TeamId       *string `json:"team_id"`

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
			ModelId:   dep.ModelId,
			ModelName: dep.Model.Name,
			Type:      dep.Model.Type,
			Username:  dep.Model.User.Username,
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

func (s *ModelService) Details(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

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

	writeJsonResponse(w, info)
}

func (s *ModelService) List(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, s.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var models []schema.Model
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Preload("Dependencies").Preload("Attributes").Preload("User").Find(&models)
	} else {
		userTeams := make([]string, 0, len(user.Teams))
		for _, t := range user.Teams {
			userTeams = append(userTeams, t.TeamId)
		}
		result = s.db.
			Preload("Dependencies").Preload("Attributes").Preload("User").
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

	writeJsonResponse(w, infos)
}

func countTrainingChildModels(db *gorm.DB, modelId string) (int64, error) {
	var childModels int64
	result := db.
		Where("base_model_id = ?", modelId).
		Where("train_status IN ?", []string{schema.NotStarted, schema.Starting, schema.InProgress}).
		Count(&childModels)

	if result.Error != nil {
		return 0, schema.NewDbError("counting child models", result.Error)
	}
	return childModels, nil
}

func (s *ModelService) Delete(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
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
			err = s.nomad.StopJob(nomad.TrainJobName(model))
			if err != nil {
				return err
			}
		}

		if model.DeployStatus == schema.Starting || model.DeployStatus == schema.InProgress || model.DeployStatus == schema.Complete {
			err = s.nomad.StopJob(nomad.DeployJobName(model))
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

	w.WriteHeader(http.StatusOK)
}

func (s *ModelService) SaveDeployed(w http.ResponseWriter, r *http.Request) {

}

func (s *ModelService) UploadToken(w http.ResponseWriter, r *http.Request) {

}

func (s *ModelService) UploadChunk(w http.ResponseWriter, r *http.Request) {

}

func (s *ModelService) UploadCommit(w http.ResponseWriter, r *http.Request) {

}

func (s *ModelService) Download(w http.ResponseWriter, r *http.Request) {

}

func (s *ModelService) UpdateAccess(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") || !params.Has("new_access") {
		http.Error(w, "'model_id' or 'new_access' query parameters missing", http.StatusBadRequest)
		return
	}
	modelId, newAccess := params.Get("model_id"), params.Get("new_access")

	if err := schema.CheckValidAccess(newAccess); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		// TODO(Nicholas) should this just be update? need to have error message if model not found
		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			return err
		}

		model.Access = newAccess

		result := txn.Save(&model)
		if result.Error != nil {
			return schema.NewDbError("updating model access", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error updating model access: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ModelService) UpdateDefaultPermission(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") || !params.Has("new_permission") {
		http.Error(w, "'model_id' or 'new_permission' query parameters missing", http.StatusBadRequest)
		return
	}
	modelId, newPermission := params.Get("model_id"), params.Get("new_permission")

	if err := schema.CheckValidPermission(newPermission); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		// TODO(Nicholas) should this just be update? need to have error message if model not found
		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			return err
		}

		model.DefaultPermission = newPermission

		result := txn.Save(&model)
		if result.Error != nil {
			return schema.NewDbError("updating model default permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error updating model default permission: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
