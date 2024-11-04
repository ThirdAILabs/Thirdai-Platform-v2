package routers

import (
	"log"
	"thirdai_platform/model_bazaar/auth"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type ModelBazaar struct {
	user UserRouter
	team TeamRouter
}

func NewModelBazaar(db *gorm.DB) ModelBazaar {
	userAuth := auth.NewJwtManager()

	return ModelBazaar{
		user: UserRouter{db: db, userAuth: userAuth},
		team: TeamRouter{db: db, userAuth: userAuth},
	}
}

func (m *ModelBazaar) Routes() chi.Router {
	r := chi.NewRouter()

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
