package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/jobs"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/model_bazaar/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeployService struct {
	db                 *gorm.DB
	orchestratorClient orchestrator.Client
	storage            storage.Storage

	userAuth auth.IdentityProvider
	jobAuth  *auth.JwtManager

	license   *licensing.LicenseVerifier
	variables Variables
}

func (s *DeployService) Routes() chi.Router {
	r := chi.NewRouter()

	eitherOrMiddleware := eitherUserOrApiKeyAuthMiddleware(s.db, s.userAuth.AuthMiddleware())
	r.Route("/{model_id}", func(r chi.Router) {
		r.Use(eitherOrMiddleware)

		r.Group(func(r chi.Router) {
			r.Use(auth.ModelPermissionOnly(s.db, auth.OwnerPermission))

			r.With(checkSufficientStorage(s.storage)).Post("/", s.Start)
			r.Delete("/", s.Stop)
		})

		r.Group(func(r chi.Router) {
			r.Use(auth.ModelPermissionOnly(s.db, auth.ReadPermission))

			r.Get("/status", s.GetStatus)
			r.Get("/logs", s.Logs)

			r.Post("/save", s.SaveDeployed)
		})

	})

	r.Group(func(r chi.Router) {
		r.Use(s.jobAuth.Verifier())
		r.Use(s.jobAuth.Authenticator())

		r.Get("/status-internal", s.GetStatusInternal)
		r.Post("/update-status", s.UpdateStatus)
		r.Post("/log", s.JobLog)
	})

	return r
}

func getDeploymentMemory(modelId uuid.UUID, userSpecified int, attrs map[string]string) int {
	if userSpecified > 500 {
		slog.Info("using user specified memory for deployment memory", "model_id", modelId, "memory", userSpecified)
		return userSpecified
	} else if userSpecified > 0 {
		slog.Error("user specified memory is to low", "model_id", modelId, "memory", userSpecified)
	}

	metadataJson, ok := attrs["metadata"]
	if ok {
		var metadata map[string]interface{}
		err := json.Unmarshal([]byte(metadataJson), &metadata)
		if err != nil {
			slog.Error("error parsing model metadata", "error", err)
		} else {
			if sizeInMemoryAny, ok := metadata["size_in_memory"]; ok {
				if sizeInMemoryStr, ok := sizeInMemoryAny.(string); ok {
					if sizeInMemory, err := strconv.Atoi(sizeInMemoryStr); err == nil {
						mem := sizeInMemory/1000000 + 1000
						slog.Info("using memory metadata for deployment memory", "model_id", modelId, "memory", mem)
						return mem
					}
				}
			}
		}
	}

	slog.Info("using default for deployment memory", "model_id", modelId, "memory", 1000)
	return 1000
}

func (s *DeployService) deployModel(modelId uuid.UUID, user schema.User, autoscaling bool, autoscalingMin, autoscalingMax int, memory int, deploymentName string) error {
	slog.Info("deploying model", "model_id", modelId, "autoscaling", autoscaling, "autoscalingMax", autoscalingMax, "memory", memory, "deployment_name", deploymentName)

	requiresOnPremLlm := false

	var nomadErr error = nil
	err := s.db.Transaction(func(txn *gorm.DB) error {
		perm, err := auth.GetModelPermissions(modelId, user, txn)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			slog.Error("unable to retrieve permissions for model", "model_id", modelId, "error", err)
			return CodedError(err, http.StatusInternalServerError)
		}
		if perm < auth.OwnerPermission {
			return CodedError(fmt.Errorf("user %v does not have permission to deploy model %v", user.Id, modelId), http.StatusForbidden)
		}

		model, err := schema.GetModel(modelId, txn, false, true, false)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		if model.TrainStatus != schema.Complete {
			return CodedError(fmt.Errorf("cannot deploy %v since it has train status %v", model.Id, model.TrainStatus), http.StatusUnprocessableEntity)
		}

		if model.DeployStatus == schema.Starting || model.DeployStatus == schema.InProgress || model.DeployStatus == schema.Complete {
			return nil
		}

		attrs := model.GetAttributes()

		memory := getDeploymentMemory(modelId, memory, attrs)
		resources := orchestrator.Resources{
			AllocationCores:     2,
			AllocationMhz:       2400,
			AllocationMemory:    memory,
			AllocationMemoryMax: 4 * memory,
		}

		license, err := verifyLicenseForNewJob(s.orchestratorClient, s.license, resources.AllocationMhz)
		if err != nil {
			return CodedError(err, GetResponseCode(err))
		}

		token, err := s.jobAuth.CreateModelJwt(modelId, time.Hour*1000*24)
		if err != nil {
			slog.Error("job token creation failed", "model_id", modelId, "error", err)
			return CodedError(errors.New("error setting up model deployment"), http.StatusInternalServerError)
		}

		if llm, hasLlm := attrs["llm_provider"]; hasLlm {
			if llm == "on-prem" {
				requiresOnPremLlm = true
			} else {
				attrs["genai_key"] = s.variables.LlmProviders[llm]
			}
		}

		var hostDir string
		if s.variables.BackendDriver.DriverType() == "local" {
			hostDir = filepath.Join(s.variables.ShareDir, "host_dir")
		} else {
			hostDir = filepath.Join("/thirdai_platform", "host_dir")
		}

		config := config.DeployConfig{
			ModelId:             model.Id,
			UserId:              user.Id,
			ModelType:           model.Type,
			ModelBazaarDir:      s.storage.Location(),
			HostDir:             hostDir,
			ModelBazaarEndpoint: s.variables.ModelBazaarEndpoint,
			LicenseKey:          license,
			JobAuthToken:        token,
			Autoscaling:         autoscaling,
			Options:             attrs,
		}

		configPath, err := saveConfig(config.ModelId, "deploy", config, s.storage)
		if err != nil {
			return CodedError(errors.New("error creating model deployment config"), http.StatusInternalServerError)
		}

		nomadErr = s.orchestratorClient.StartJob(
			orchestrator.DeployJob{
				JobName:            model.DeployJobName(),
				ModelId:            model.Id.String(),
				ConfigPath:         configPath,
				DeploymentName:     deploymentName,
				AutoscalingEnabled: autoscaling,
				AutoscalingMin:     autoscalingMin,
				AutoscalingMax:     autoscalingMax,
				Driver:             s.variables.BackendDriver,
				Resources:          resources,
				CloudCredentials:   s.variables.CloudCredentials,
				JobToken:           uuid.New().String(),
				IsKE:               model.Type == schema.KnowledgeExtraction,
				IngressHostname:    s.orchestratorClient.IngressHostname(),
				Platform:           s.variables.Platform,
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
			slog.Error("sql error updating deploy status on job start", "model_id", model.Id, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		return CodedError(fmt.Errorf("error starting model deployment: %w", err), GetResponseCode(err))
	}

	if nomadErr != nil {
		return CodedError(errors.New("error starting model deployment on nomad"), http.StatusInternalServerError)
	}

	if requiresOnPremLlm {
		err := jobs.StartOnPremGenerationJobDefaultArgs(s.orchestratorClient, s.storage, s.variables.DockerEnv(), s.variables.ShareDir, s.variables.Platform)
		if err != nil {
			slog.Error("error starting on-prem-generation job", "error", err)
			return CodedError(errors.New("unable to start on prem generation job"), http.StatusInternalServerError)
		}
	}

	slog.Info("model deployed successfully", "model_id", modelId)

	return nil
}

type startRequest struct {
	DeploymentName string `json:"deployment_name"`
	Autoscaling    bool   `json:"autoscaling_enabled"`
	AutoscalingMin int    `json:"autoscaling_min"`
	AutoscalingMax int    `json:"autoscaling_max"`
	Memory         int    `json:"memory"`
}

func (s *DeployService) Start(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var params startRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	params.AutoscalingMin = max(params.AutoscalingMin, 1)
	params.AutoscalingMax = max(params.AutoscalingMax, 1)

	deps, err := listModelDependencies(modelId, s.db)
	if err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	for _, dep := range deps {
		name := ""
		if dep.Id == modelId {
			name = params.DeploymentName
		}
		err := s.deployModel(dep.Id, user, params.Autoscaling, params.AutoscalingMin, params.AutoscalingMax, params.Memory, name)
		if err != nil {
			http.Error(w, err.Error(), GetResponseCode(err))
			return
		}
	}

	utils.WriteSuccess(w)
}

func (s *DeployService) Stop(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("stopping deployment for model", "model_id", modelId)

	err = s.db.Transaction(func(txn *gorm.DB) error {
		usedBy, err := countDownstreamModels(modelId, txn, true)
		if err != nil {
			return fmt.Errorf("error checking if model is a dependend of other models: %w", err)
		}
		if usedBy != 0 {
			return CodedError(fmt.Errorf("cannot stop deployment for model %v since it is used as a dependency by %d other active models", modelId, usedBy), http.StatusUnprocessableEntity)
		}

		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		err = s.orchestratorClient.StopJob(model.DeployJobName())
		if err != nil {
			slog.Error("error stopping deployment", "error", err)
			return CodedError(errors.New("error stopping deployment job"), http.StatusInternalServerError)
		}

		result := txn.Model(&model).Update("deploy_status", schema.Stopped)
		if result.Error != nil {
			slog.Error("sql error updating deploy status on job stop", "model_id", model.Id, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error stopping model deployment: %v", err), GetResponseCode(err))
		return
	}

	slog.Info("model stopped successfully", "model_id", modelId)

	utils.WriteSuccess(w)
}

func (s *DeployService) GetStatusInternal(w http.ResponseWriter, r *http.Request) {
	modelId, err := auth.ModelIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	getStatusHandler(w, modelId, s.db, "deploy")
}

func (s *DeployService) GetStatus(w http.ResponseWriter, r *http.Request) {
	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	getStatusHandler(w, modelId, s.db, "deploy")
}

func (s *DeployService) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	updateStatusHandler(w, r, s.db, "deploy")
}

func (s *DeployService) Logs(w http.ResponseWriter, r *http.Request) {
	getLogsHandler(w, r, s.db, s.orchestratorClient, "deploy")
}

func (s *DeployService) JobLog(w http.ResponseWriter, r *http.Request) {
	jobLogHandler(w, r, s.db, "deploy")
}

type saveDeployedRequest struct {
	ModelName string `json:"model_name"`
}

type saveDeployedResponse struct {
	ModelId     uuid.UUID `json:"model_id"`
	UpdateToken string    `json:"update_token"`
}

func (s *DeployService) SaveDeployed(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving user id from request: %v", err), http.StatusInternalServerError)
		return
	}

	var params saveDeployedRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	baseModelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newModelId := uuid.New()

	err = s.db.Transaction(func(txn *gorm.DB) error {
		baseModel, err := schema.GetModel(baseModelId, txn, true, true, false)
		if err != nil {
			if errors.Is(err, schema.ErrModelNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		err = checkForDuplicateModel(txn, params.ModelName, user.Id)
		if err != nil {
			slog.Info("unable to save deployed model: duplicate model name", "base_model_id", baseModel.Id, "model_name", params.ModelName)
			return err
		}

		model := newModel(newModelId, params.ModelName, baseModel.Type, &baseModel.Id, user.Id)

		return saveModel(txn, model, user)
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error saving new deployed model: %v", err), GetResponseCode(err))
		return
	}

	updateToken, err := s.jobAuth.CreateModelJwt(newModelId, time.Hour)
	if err != nil {
		slog.Error("error generating update jwt for save deployed", "error", err)
		http.Error(w, "error creating save token", http.StatusInternalServerError)
		return
	}

	utils.WriteJsonResponse(w, saveDeployedResponse{ModelId: newModelId, UpdateToken: updateToken})
}
