package auth

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/schema"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func ExpectAdmin(userId string, db *gorm.DB) error {
	user, err := schema.GetUser(userId, db, false)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return fmt.Errorf("user %v is not an admin", userId)
	}
	return nil
}

func AdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			userId, err := UserIdFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			err = ExpectAdmin(userId, db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
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

			userId, err := UserIdFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			user, err := schema.GetUser(userId, db, false)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin && !isTeamAdmin(teamId, userId, db) {
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

			userId, err := UserIdFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			user, err := schema.GetUser(userId, db, false)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin && !isTeamMember(teamId, userId, db) {
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

func GetModelPermissions(modelId, userId string, db *gorm.DB) (modelPermissions, error) {
	user, err := schema.GetUser(userId, db, false)
	if err != nil {
		return NoPermission, err
	}

	if user.IsAdmin {
		return OwnerPermission, nil
	}

	model, err := schema.GetModel(modelId, db, false, false, false)
	if err != nil {
		return NoPermission, err
	}

	if model.UserId == userId {
		return OwnerPermission, nil
	}

	if model.Access == schema.Public {
		if model.DefaultPermission == schema.WritePerm {
			return WritePermission, nil
		}
		return ReadPermission, nil
	}

	if model.Access == schema.Protected && model.TeamId != nil {
		userTeam, err := schema.GetUserTeam(*model.TeamId, userId, db)
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

			userId, err := UserIdFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			permission, err := GetModelPermissions(modelId, userId, db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if permission >= minPermission {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, fmt.Sprintf("user %v does not have owner permission for model %v", userId, modelId), http.StatusUnauthorized)
		}
		return http.HandlerFunc(hfn)
	}
}
