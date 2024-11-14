package services

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"time"

	"github.com/go-chi/chi/v5"
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
		r.Post("/log", s.JobLog)
	})

	return r
}

func (s *DeployService) deployModel(modelId, userId string, autoscalingEnabled bool, autoscalingMax int, memory int, deploymentName string) error {
	slog.Info("deploying model", "model_id", modelId)

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

		if model.TrainStatus != schema.Complete {
			return fmt.Errorf("cannot deploy %v since it has train status %v", model.Id, model.TrainStatus)
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

		token, err := s.jobAuth.CreateToken("model_id", modelId, time.Hour*1000*24)
		if err != nil {
			return fmt.Errorf("error creating job token: %v", err)
		}

		attrs := model.GetAttributes()

		config := config.DeployConfig{
			ModelId:             model.Id,
			ModelBazaarDir:      s.storage.Location(),
			ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
			LicenseKey:          license,
			JobAuthToken:        token,
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

	slog.Info("model deployed successfully", "model_id", modelId)

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

	params.Memory = max(params.Memory, 1000)
	params.AutoscalingMax = max(params.AutoscalingMax, 1)

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

	writeSuccess(w)
}

func (s *DeployService) Stop(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

	slog.Info("stopping deployment for model", "model_id", modelId)

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

		result := txn.Model(&model).Update("deploy_status", schema.Stopped)
		if result.Error != nil {
			return schema.NewDbError("updating deploy status for stopped model", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error stopping model deployment: %v", err), http.StatusBadRequest)
		return
	}

	slog.Info("model stopped successfully", "model_id", modelId)

	writeSuccess(w)
}

func (s *DeployService) GetStatus(w http.ResponseWriter, r *http.Request) {
	getStatusHandler(w, r, s.db, "deploy")
}

func (s *DeployService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	updateStatusHandler(w, r, s.db, "deploy")
}

func (s *DeployService) Logs(w http.ResponseWriter, r *http.Request) {
	getLogsHandler(w, r, s.db, s.nomad, "deploy")
}

func (s *DeployService) JobLog(w http.ResponseWriter, r *http.Request) {
	jobLogHandler(w, r, s.db, "deploy")
}
