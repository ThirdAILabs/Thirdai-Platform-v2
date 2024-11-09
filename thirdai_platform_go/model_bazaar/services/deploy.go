package services

import (
	"errors"
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

type DeployService struct {
	db      *gorm.DB
	nomad   nomad.NomadClient
	storage storage.Storage

	userAuth *auth.JwtManager
	jobAuth  *auth.JwtManager

	license   *licensing.LicenseVerifier
	variables Variables
}

func (s *DeployService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())

		r.Post("/start", s.Start)
		r.Get("/permissions", s.Permissions)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.ReadPermission))

		r.Get("/status", s.GetStatus)
		r.Get("/logs", s.Logs)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.OwnerPermission))

		r.Post("/stop", s.Stop)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.jobAuth.Verifier())
		r.Use(s.jobAuth.Authenticator())

		r.Post("/update-status", s.UpdateStatus)
	})

	return r
}

type permissionsResponse struct {
	Read     bool      `json:"read"`
	Write    bool      `json:"write"`
	Override bool      `json:"override"`
	Exp      time.Time `json:"exp"`
}

func (s *DeployService) Permissions(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") || !params.Has("user_id") {
		http.Error(w, "'model_id' or 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	modelId, userId := params.Get("model_id"), params.Get("user_id")

	permission, err := auth.GetModelPermissions(modelId, userId, s.db)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving model permissions: %v", err), http.StatusBadRequest)
		return
	}

	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving access token: %v", err), http.StatusBadRequest)
		return
	}

	res := permissionsResponse{
		Read:     permission >= auth.ReadPermission,
		Write:    permission >= auth.WritePermission,
		Override: permission >= auth.OwnerPermission,
		Exp:      token.Expiration(),
	}
	writeJsonResponse(w, res)
}

func (s *DeployService) deployModel(modelId, userId string, autoscalingEnabled bool, autoscalingMax int, memory int, deploymentName string) error {
	var nomadErr error = nil
	err := s.db.Transaction(func(txn *gorm.DB) error {
		perm, err := auth.GetModelPermissions(modelId, userId, txn)
		if err != nil {
			return fmt.Errorf("unable to retrieve permission for model: %w", err)
		}
		if perm < auth.OwnerPermission {
			return fmt.Errorf("user %v does not have permission to deploy model %v", userId, modelId)
		}

		model, err := schema.GetModel(modelId, txn, false, true, false)
		if err != nil {
			return err
		}

		if model.DeployStatus == schema.Starting || model.DeployStatus == schema.InProgress || model.DeployStatus == schema.Complete {
			return nil
		}

		// TODO(Nicholas) : autotune memory from metadata if present
		resources := nomad.Resources{
			AllocationMhz:       2400,
			AllocationMemory:    memory,
			AllocationMemoryMax: 4 * memory,
		}

		license, err := verifyLicenseForNewJob(s.nomad, s.license, resources.AllocationMhz)
		if err != nil {
			return err
		}

		attrs := model.GetAttributes()

		config := config.DeployConfig{
			ModelId:             model.Id,
			ModelBazaarDir:      s.storage.Location(),
			ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
			LicenseKey:          license,
			AutoscalingEnabled:  autoscalingEnabled,
			Options:             attrs,
		}

		configPath, err := saveConfig(config.ModelId, "deploy", config, s.storage)
		if err != nil {
			return err
		}

		nomadErr = s.nomad.StartJob(
			nomad.DeployJob{
				JobName:            model.DeployJobName(),
				ModelId:            model.Id,
				ConfigPath:         configPath,
				DeploymentName:     deploymentName,
				AutoscalingEnabled: autoscalingEnabled,
				AutoscalingMax:     autoscalingMax,
				Driver:             s.variables.Driver,
				Resources:          resources,
			},
		)
		var newStatus string
		if nomadErr != nil {
			newStatus = schema.Failed
		} else {
			newStatus = schema.Starting
		}

		result := txn.Model(&model).Update("deploy_status", newStatus)
		if result.Error != nil {
			return schema.NewDbError("updating model deploy status", result.Error)
		}

		return nil
	})

	// TODO(nicholas): start on prem llm if needed

	if jerr := errors.Join(err, nomadErr); jerr != nil {
		return fmt.Errorf("error starting deployment for model %v: %w", modelId, jerr)
	}

	return nil
}

type startRequest struct {
	ModelId            string `json:"model_id"`
	DeploymentName     string `json:"deployment_name"`
	AutoScalingEnabled bool   `json:"autoscaling_enabled"`
	AutoscalingMax     int    `json:"autoscaling_max"`
	Memory             int    `json:"memory"`
}

func (s *DeployService) Start(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params startRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	deps, err := listModelDependencies(params.ModelId, s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, dep := range deps {
		name := ""
		if dep.Id == params.ModelId {
			name = params.DeploymentName
		}
		err := s.deployModel(dep.Id, userId, params.AutoScalingEnabled, params.AutoscalingMax, params.Memory, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *DeployService) Stop(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
		usedBy, err := countDownstreamModels(modelId, txn, true)
		if err != nil {
			return err
		}
		if usedBy != 0 {
			return fmt.Errorf("cannot stop deployment for model %v since it is used as a dependency by %d other active models", modelId, usedBy)
		}

		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			return err
		}

		err = s.nomad.StopJob(model.DeployJobName())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error stopping model deployment: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *DeployService) GetStatus(w http.ResponseWriter, r *http.Request) {
	getStatusHandler(w, r, s.db, false)
}

func (s *DeployService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	updateStatusHandler(w, r, s.db, false)
}

func (s *DeployService) Logs(w http.ResponseWriter, r *http.Request) {
	getLogsHandler(w, r, s.db, s.nomad, true)
}
