package routers

import (
	"thirdai_platform/model_bazaar/nomad"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type DeployRouter struct {
	db    *gorm.DB
	nomad nomad.NomadClient
}

func (d *DeployRouter) Routes() chi.Router {
	return nil
}
