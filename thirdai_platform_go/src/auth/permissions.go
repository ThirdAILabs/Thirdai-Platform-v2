package auth

import (
	"fmt"
	"net/http"
	"thirdai_platform/src/schema"

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

func isTeamAdmin(teamId string, user *schema.User) bool {
	for _, role := range user.Teams {
		if role.TeamId == teamId && role.IsTeamAdmin {
			return true
		}
	}
	return false
}

func AdminOrTeamAdminOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()
			if !params.Has("team_id") {
				http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
				return
			}
			teamId := params.Get("team_id")

			userId, err := UserIdFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			user, err := schema.GetUser(userId, db, true)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin && !isTeamAdmin(teamId, &user) {
				http.Error(w, "user must be admin or team admin to access endpoint", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func ModelOwnerOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()
			if !params.Has("model_id") {
				http.Error(w, "'model_id' query parameter missing", http.StatusBadRequest)
				return
			}
			modelId := params.Get("model_id")

			userId, err := UserIdFromContext(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			user, err := schema.GetUser(userId, db, true)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			model, err := schema.GetModel(modelId, db, false, false)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !user.IsAdmin && model.UserId != userId && !(model.TeamId != nil && model.Access == schema.Protected && isTeamAdmin(*model.TeamId, &user)) {
				http.Error(w, "user must be model owner to access endpoint", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
