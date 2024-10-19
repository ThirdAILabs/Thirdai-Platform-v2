package routers

import (
	"thirdai_platform/src/nomad"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type DeployRouter struct {
	db    *gorm.DB
	nomad *nomad.Client
}

func (d *DeployRouter) Routes() chi.Router {
	return nil
}
