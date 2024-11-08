package services

import (
	"log"
	"thirdai_platform/model_bazaar/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type ModelBazaar struct {
	user UserService
	team TeamService
}

func NewModelBazaar(db *gorm.DB) ModelBazaar {
	userAuth := auth.NewJwtManager()

	return ModelBazaar{
		user: UserService{db: db, userAuth: userAuth},
		team: TeamService{db: db, userAuth: userAuth},
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
