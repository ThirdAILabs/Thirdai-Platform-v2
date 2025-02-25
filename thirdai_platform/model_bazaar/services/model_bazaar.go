package services

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/licensing"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/storage"
	"thirdai_platform/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type ModelBazaar struct {
	user      UserService
	team      TeamService
	model     ModelService
	train     TrainService
	deploy    DeployService
	telemetry TelemetryService
	workflow  WorkflowService
	recovery  RecoveryService

	db                 *gorm.DB
	orchestratorClient orchestrator.Client
	stop               chan bool
}

func NewModelBazaar(
	db *gorm.DB, orchestratorClient orchestrator.Client, storage storage.Storage, license *licensing.LicenseVerifier, userAuth auth.IdentityProvider, variables Variables, secret []byte,
) ModelBazaar {
	jobAuth := auth.NewJwtManager(slices.Concat(secret, []byte("job")))

	return ModelBazaar{
		user: UserService{db: db, userAuth: userAuth},
		team: TeamService{db: db, userAuth: userAuth},
		model: ModelService{
			db:                 db,
			orchestratorClient: orchestratorClient,
			storage:            storage,
			userAuth:           userAuth,
			uploadSessionAuth:  auth.NewJwtManager(slices.Concat(secret, []byte("upload"))),
		},
		train: TrainService{
			db:                 db,
			orchestratorClient: orchestratorClient,
			storage:            storage,
			userAuth:           userAuth,
			jobAuth:            jobAuth,
			license:            license,
			variables:          variables,
		},
		deploy: DeployService{
			db:                 db,
			orchestratorClient: orchestratorClient,
			storage:            storage,
			userAuth:           userAuth,
			jobAuth:            jobAuth,
			license:            license,
			variables:          variables,
		},
		telemetry: TelemetryService{
			orchestratorClient: orchestratorClient,
			variables:          variables,
		},
		workflow: WorkflowService{
			db:       db,
			storage:  storage,
			userAuth: userAuth,
		},
		recovery: RecoveryService{
			db:                 db,
			storage:            storage,
			orchestratorClient: orchestratorClient,
			userAuth:           userAuth,
			variables:          variables,
		},
		db:                 db,
		orchestratorClient: orchestratorClient,
		stop:               make(chan bool, 1),
	}
}

func (m *ModelBazaar) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(os.Stderr, "", log.LstdFlags), NoColor: false,
	}))

	r.Mount("/user", m.user.Routes())
	r.Mount("/team", m.team.Routes())
	r.Mount("/model", m.model.Routes())
	r.Mount("/train", m.train.Routes())
	r.Mount("/deploy", m.deploy.Routes())
	r.Mount("/telemetry", m.telemetry.Routes())
	r.Mount("/workflow", m.workflow.Routes())
	r.Mount("/recovery", m.recovery.Routes())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccess(w)
	})

	return r
}

func (m *ModelBazaar) syncTrainStatus(model *schema.Model) {
	if model.TrainStatus != schema.Starting && model.TrainStatus != schema.InProgress {
		return
	}
	jobInfo, err := m.orchestratorClient.JobInfo(model.TrainJobName())
	jobNotFound := errors.Is(err, orchestrator.ErrJobNotFound)

	if err != nil && !jobNotFound {
		slog.Error("status sync: train job info", "error", err)
		return
	}

	if jobInfo.Status == "dead" || jobNotFound {
		result := m.db.Model(model).Where("train_status = ?", model.TrainStatus).Update("train_status", schema.Failed)
		if result.Error != nil {
			slog.Error("status sync: sql error updating train status for failed training", "model_id", model.Id, "error", result.Error)
			return
		}
		slog.Info("status sync: updated train status to failed", "model_id", model.Id)
	}
}

func (m *ModelBazaar) syncDeployStatus(model *schema.Model) {
	if model.DeployStatus != schema.Starting && model.DeployStatus != schema.InProgress && model.DeployStatus != schema.Complete {
		return
	}

	jobInfo, err := m.orchestratorClient.JobInfo(model.DeployJobName())
	jobNotFound := errors.Is(err, orchestrator.ErrJobNotFound)

	if err != nil && !jobNotFound {
		slog.Error("status sync: deploy job info", "error", err)
		return
	}

	if jobInfo.Status == "dead" || jobNotFound {
		result := m.db.Model(model).Where("deploy_status = ?", model.DeployStatus).Update("deploy_status", schema.Failed)
		if result.Error != nil {
			slog.Error("status sync: sql error updating deploy status for failed deployment", "model_id", model.Id, "error", result.Error)
			return
		}

		slog.Info("status sync: updated deploy status to failed", "model_id", model.Id)
	}
}

func (m *ModelBazaar) statusSync() {
	var models []schema.Model

	result := m.db.
		Where("train_status IN ?", []string{schema.Starting, schema.InProgress}).
		Or("deploy_status IN ?", []string{schema.Starting, schema.InProgress, schema.Complete}).
		Find(&models)

	if result.Error != nil {
		slog.Error("status sync: sql error querying active models", "error", result.Error)
		return
	}

	for _, model := range models {
		m.syncTrainStatus(&model)
		m.syncDeployStatus(&model)
	}
}

func (m *ModelBazaar) JobStatusSync(interval time.Duration) {
	slog.Info("status sync: starting")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.statusSync()
		case <-m.stop:
			slog.Info("status sync: process stopped")
			return
		}
	}
}

func (m *ModelBazaar) StopJobStatusSync() {
	close(m.stop)
}
