package services

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type ModelBazaar struct {
	user UserService
	team TeamService

	db    *gorm.DB
	nomad nomad.NomadClient
	stop  chan bool
}

func NewModelBazaar(db *gorm.DB, nomad nomad.NomadClient) ModelBazaar {
	userAuth := auth.NewJwtManager()

	return ModelBazaar{
		user:  UserService{db: db, userAuth: userAuth},
		team:  TeamService{db: db, userAuth: userAuth},
		db:    db,
		nomad: nomad,
		stop:  make(chan bool, 1),
	}
}

func (m *ModelBazaar) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Mount("/user", m.user.Routes())
	r.Mount("/team", m.team.Routes())

	return r
}

func (m *ModelBazaar) InitAdmin(username, email, password string) {
	_, err := m.user.CreateUser(username, email, password, true)
	if err != nil {
		log.Panicf("error initializing admin at startup: %v", err)
	}
}

func (m *ModelBazaar) syncTrainStatus(model *schema.Model) {
	if model.TrainStatus != schema.Starting && model.TrainStatus != schema.InProgress {
		return
	}
	jobInfo, err := m.nomad.JobInfo(model.TrainJobName())
	jobNotFound := errors.Is(err, nomad.ErrJobNotFound)

	if err != nil && !jobNotFound {
		slog.Error(fmt.Sprintf("status sync: %v", err))
		return
	}

	if jobInfo.Status == "dead" || jobNotFound {
		result := m.db.Model(model).Where("train_status = ?", model.TrainStatus).Update("train_status", schema.Failed)
		if result.Error != nil {
			err := schema.NewDbError("updating train status for failed model", result.Error)
			slog.Error(fmt.Sprintf("status sync: %v", err))
			return
		}
	}
}

func (m *ModelBazaar) syncDeployStatus(model *schema.Model) {
	if model.DeployStatus != schema.Starting && model.DeployStatus != schema.InProgress && model.DeployStatus != schema.Complete {
		return
	}

	jobInfo, err := m.nomad.JobInfo(model.DeployJobName())
	jobNotFound := errors.Is(err, nomad.ErrJobNotFound)

	if err != nil && !jobNotFound {
		slog.Error(fmt.Sprintf("status sync: %v", err))
		return
	}

	if jobInfo.Status == "dead" || jobNotFound {
		result := m.db.Model(model).Where("deploy_status = ?", model.DeployStatus).Update("deploy_status", schema.Failed)
		if result.Error != nil {
			err := schema.NewDbError("updating deploy status for failed model", result.Error)
			slog.Error(fmt.Sprintf("status sync: %v", err))
			return
		}
	}
}

func (m *ModelBazaar) statusSync() {
	var models []schema.Model

	result := m.db.
		Where("train_status IN ?", []string{schema.Starting, schema.InProgress}).
		Or("deploy_status IN ?", []string{schema.Starting, schema.InProgress, schema.Complete}).
		Find(&models)

	if result.Error != nil {
		err := schema.NewDbError("listing active models", result.Error)
		slog.Error(fmt.Sprintf("status sync: %v", err))
		return
	}

	for _, model := range models {
		m.syncTrainStatus(&model)
		m.syncDeployStatus(&model)
	}
}

func (m *ModelBazaar) StartStatusSync() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.statusSync()
		case <-m.stop:
			slog.Info("status sync process stopped")
			return
		}
	}
}

func (m *ModelBazaar) StopStatusSync() {
	close(m.stop)
}
