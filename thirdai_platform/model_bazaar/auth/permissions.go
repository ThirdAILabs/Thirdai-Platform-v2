package auth

import (
	"errors"
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !user.IsAdmin {
				http.Error(w, fmt.Sprintf("user %v is not an admin", user.Id), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func isTeamAdmin(teamId, userId uuid.UUID, db *gorm.DB) (bool, error) {
	userTeam, err := schema.GetUserTeam(teamId, userId, db)
	if err != nil {
		if errors.Is(err, schema.ErrUserTeamNotFound) {
			return false, nil
		}
		return false, err
	}

	return userTeam.IsTeamAdmin, nil
}

func AdminOrTeamAdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			teamId, err := utils.URLParamUUID(r, "team_id")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			isAdmin, err := isTeamAdmin(teamId, user.Id, db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !user.IsAdmin && !isAdmin {
				http.Error(w, "user must be admin or team admin to access endpoint", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func isTeamMember(teamId, userId uuid.UUID, db *gorm.DB) (bool, error) {
	_, err := schema.GetUserTeam(teamId, userId, db)
	if err != nil {
		if errors.Is(err, schema.ErrUserTeamNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func TeamMemberOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			teamId, err := utils.URLParamUUID(r, "team_id")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			isMember, err := isTeamMember(teamId, user.Id, db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !user.IsAdmin && !isMember {
				http.Error(w, "user must be team member to access endpoint", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

type modelPermission int // Private so that no other permissions can be defined

const (
	NoPermission    modelPermission = 0
	ReadPermission  modelPermission = 1
	WritePermission modelPermission = 2
	OwnerPermission modelPermission = 3
)

func modelPermissionToString(perm modelPermission) string {
	switch perm {
	case NoPermission:
		return "None"
	case ReadPermission:
		return "Read"
	case WritePermission:
		return "Write"
	case OwnerPermission:
		return "Owner"
	default:
		return "invalid permission"
	}
}

func GetModelPermissions(modelId uuid.UUID, user schema.User, db *gorm.DB) (modelPermission, error) {
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
			if errors.Is(err, schema.ErrUserTeamNotFound) {
				return NoPermission, nil
			}
			return NoPermission, err
		}
		if userTeam.IsTeamAdmin {
			return OwnerPermission, nil
		}
		if model.DefaultPermission == schema.WritePerm {
			return WritePermission, nil
		}
		return ReadPermission, nil
	}

	return NoPermission, nil
}

func ModelPermissionOnly(db *gorm.DB, minPermission modelPermission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			modelId, err := utils.URLParamUUID(r, "model_id")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			user, err := UserFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			permission, err := GetModelPermissions(modelId, user, db)
			if err != nil {
				if errors.Is(err, schema.ErrModelNotFound) {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if permission >= minPermission {
				next.ServeHTTP(w, r)
				return
			}

			required, actual := modelPermissionToString(minPermission), modelPermissionToString(permission)
			http.Error(w, fmt.Sprintf("user %v does not have required permission for model %v (required=%v, actual=%v)", user.Id, modelId, required, actual), http.StatusForbidden)
		}
		return http.HandlerFunc(hfn)
	}
}
