package routers

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func Service(db *gorm.DB) chi.Router {
	r := chi.NewRouter()

	userRouter := NewUserRouter(db)
	r.Mount("/user", userRouter.Routes())

	return r
}
