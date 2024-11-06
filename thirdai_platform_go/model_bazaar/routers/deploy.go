package routers

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

type DeployRouter struct {
	db    *gorm.DB
	nomad nomad.NomadClient

	userAuth *auth.JwtManager
	jobAuth  *auth.JwtManager
}

func (d *DeployRouter) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(d.userAuth.Verifier())
		r.Use(d.userAuth.Authenticator())

		r.Get("/permissions", d.Permissions)
	})

	r.Group(func(r chi.Router) {
		r.Use(d.userAuth.Verifier())
		r.Use(d.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(d.db, auth.ReadPermission))

		r.Get("/status", d.GetStatus)
		r.Get("/logs", d.Logs)
	})

	r.Group(func(r chi.Router) {
		r.Use(d.userAuth.Verifier())
		r.Use(d.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(d.db, auth.OwnerPermission))

		r.Post("/start", d.Start)
		r.Post("/stop", d.Stop)
	})

	r.Group(func(r chi.Router) {
		r.Use(d.jobAuth.Verifier())
		r.Use(d.jobAuth.Authenticator())

		r.Post("/update-status", d.UpdateStatus)
	})

	return r
}

type permissionsResponse struct {
	Read     bool      `json:"read"`
	Write    bool      `json:"write"`
	Override bool      `json:"override"`
	Exp      time.Time `json:"exp"`
}

func (d *DeployRouter) Permissions(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") || !params.Has("user_id") {
		http.Error(w, "'model_id' or 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	modelId, userId := params.Get("model_id"), params.Get("user_id")

	permission, err := auth.GetModelPermissions(modelId, userId, d.db)
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

func (d *DeployRouter) Start(w http.ResponseWriter, r *http.Request) {

}

func (d *DeployRouter) Stop(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("model_id") {
		http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
		return
	}
	modelId := params.Get("model_id")

	err := d.db.Transaction(func(db *gorm.DB) error {
		usedBy, err := countDownstreamModels(modelId, db, true)
		if err != nil {
			return err
		}
		if usedBy != 0 {
			return fmt.Errorf("cannot stop deployment for model %v since it is used as a dependency by %d other active models", modelId, usedBy)
		}

		model, err := schema.GetModel(modelId, db, false, false, false)
		if err != nil {
			return err
		}

		err = d.nomad.StopJob(nomad.DeployJobName(model))
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

func (d *DeployRouter) GetStatus(w http.ResponseWriter, r *http.Request) {
	getStatusHandler(w, r, d.db, false)
}

func (d *DeployRouter) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	updateStatusHandler(w, r, d.db, false)
}

func (d *DeployRouter) Logs(w http.ResponseWriter, r *http.Request) {

}
