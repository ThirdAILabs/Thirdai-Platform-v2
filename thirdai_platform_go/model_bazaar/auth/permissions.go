package auth

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/schema"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func AdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin {
				http.Error(w, fmt.Sprintf("user %v is not an admin", user.Id), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func isTeamAdmin(teamId, userId string, db *gorm.DB) bool {
	userTeam, err := schema.GetUserTeam(teamId, userId, db)
	if err != nil || userTeam == nil {
		return false
	}

	return userTeam.IsTeamAdmin
}

func AdminOrTeamAdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			teamId := chi.URLParam(r, "team_id")

			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin && !isTeamAdmin(teamId, user.Id, db) {
				http.Error(w, "user must be admin or team admin to access endpoint", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func isTeamMember(teamId, userId string, db *gorm.DB) bool {
	userTeam, err := schema.GetUserTeam(teamId, userId, db)
	if err != nil || userTeam == nil {
		return false
	}

	return true
}

func TeamMemberOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			teamId := chi.URLParam(r, "team_id")

			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin && !isTeamMember(teamId, user.Id, db) {
				http.Error(w, "user must be team member to access endpoint", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

type modelPermissions int // Private so that no other permissions can be defined

const (
	NoPermission    modelPermissions = 0
	ReadPermission  modelPermissions = 1
	WritePermission modelPermissions = 2
	OwnerPermission modelPermissions = 3
)

func GetModelPermissions(modelId string, user schema.User, db *gorm.DB) (modelPermissions, error) {
	if user.IsAdmin {
		return OwnerPermission, nil
	}

	model, err := schema.GetModel(modelId, db, false, false, false)
	if err != nil {
		return NoPermission, err
	}

	if model.UserId == user.Id {
		return OwnerPermission, nil
	}

	if model.Access == schema.Public {
		if model.DefaultPermission == schema.WritePerm {
			return WritePermission, nil
		}
		return ReadPermission, nil
	}

	if model.Access == schema.Protected && model.TeamId != nil {
		userTeam, err := schema.GetUserTeam(*model.TeamId, user.Id, db)
		if err != nil {
			return NoPermission, err
		}
		if userTeam != nil {
			if userTeam.IsTeamAdmin {
				return OwnerPermission, nil
			}
			if model.DefaultPermission == schema.WritePerm {
				return WritePermission, nil
			}
			return ReadPermission, nil
		}
	}

	return NoPermission, nil
}

func ModelPermissionOnly(db *gorm.DB, minPermission modelPermissions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			modelId := chi.URLParam(r, "model_id")

			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			permission, err := GetModelPermissions(modelId, user, db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if permission >= minPermission {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, fmt.Sprintf("user %v does not have owner permission for model %v", user.Id, modelId), http.StatusUnauthorized)
		}
		return http.HandlerFunc(hfn)
	}
}
