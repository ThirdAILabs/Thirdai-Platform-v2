package routers

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamRouter struct {
	db       *gorm.DB
	userAuth *auth.JwtManager
}

func (t *TeamRouter) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(t.userAuth.Verifier())
		r.Use(t.userAuth.Authenticator())
		r.Use(auth.AdminOnly(t.db))

		r.Post("/create", t.CreateTeam)
		r.Post("/delete", t.DeleteTeam)

	})

	r.Group(func(r chi.Router) {
		r.Use(t.userAuth.Verifier())
		r.Use(t.userAuth.Authenticator())
		r.Use(auth.AdminOrTeamAdminOnly(t.db))

		r.Post("/add-user", t.AddUserToTeam)
		r.Post("/remove-user", t.RemoveUserFromTeam)

		r.Post("/add-admin", t.AddTeamAdmin)
		r.Post("/remove-admin", t.RemoveTeamAdmin)

		r.Get("/users", t.TeamUsers)
		r.Get("/models", t.TeamModels)
	})

	r.Group(func(r chi.Router) {
		r.Use(t.userAuth.Verifier())
		r.Use(t.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(t.db, auth.OwnerPermission))

		r.Post("/add-model", t.AddModelToTeam)
		r.Post("/remove-model", t.RemoveModelFromTeam)
	})

	r.Group(func(r chi.Router) {
		r.Use(t.userAuth.Verifier())
		r.Use(t.userAuth.Authenticator())

		r.Get("/list", t.List)
	})

	return r
}

func (t *TeamRouter) CreateTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("name") {
		http.Error(w, "'name' query parameter missing", http.StatusBadRequest)
		return
	}
	name := params.Get("name")

	newTeam := schema.Team{Id: uuid.New().String(), Name: name}

	err := t.db.Transaction(func(db *gorm.DB) error {
		var existingTeam schema.Team
		result := db.Find(&existingTeam, "name = ?", name)
		if result.Error != nil {
			return schema.NewDbError("checking for existing team with name", result.Error)
		}
		if result.RowsAffected != 0 {
			return fmt.Errorf("team with name %v already exists", name)
		}

		result = db.Create(&newTeam)
		if result.Error != nil {
			return schema.NewDbError("creating new team entry", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error creating team: %v", err), http.StatusBadRequest)
		return
	}

	writeJsonResponse(w, map[string]string{"team_id": newTeam.Id})
}

func (t *TeamRouter) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") {
		http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
		return
	}

	team := schema.Team{Id: params.Get("team_id")}

	err := t.db.Transaction(func(db *gorm.DB) error {
		exists, err := schema.TeamExists(db, team.Id)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("team %v does not exists", team.Id)
		}

		result := db.Delete(&team)
		if result.Error != nil {
			return schema.NewDbError("deleting team", result.Error)
		}

		result = db.Model(&schema.Model{}).Where("team_id = ?", team.Id).Update("team_id", nil).Update("access", schema.Private)
		if result.Error != nil {
			return schema.NewDbError("updating access for models after deleting team", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting team: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *TeamRouter) AddUserToTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	userTeam := schema.UserTeam{UserId: userId, TeamId: teamId}

	err := t.db.Transaction(func(db *gorm.DB) error {
		teamExists, err := schema.TeamExists(db, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(db, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := db.Create(&userTeam)
		if result.Error != nil {
			return schema.NewDbError("creating new user_team entry", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding user to team: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *TeamRouter) RemoveUserFromTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	err := t.db.Transaction(func(db *gorm.DB) error {
		teamExists, err := schema.TeamExists(db, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(db, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := db.Delete(&schema.UserTeam{UserId: userId, TeamId: teamId})
		if result.Error != nil {
			return schema.NewDbError("deleting user_team entry", result.Error)
		}

		result = db.Model(&schema.Model{}).Where("team_id = ? and user_id = ?", teamId, userId).Update("team_id", nil).Update("access", schema.Private)
		if result.Error != nil {
			return schema.NewDbError("updating model permissions after removing user from team", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing user from team: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *TeamRouter) AddModelToTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("model_id") {
		http.Error(w, "'team_id' and 'model_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, modelId := params.Get("team_id"), params.Get("model_id")

	err := t.db.Transaction(func(db *gorm.DB) error {
		teamExists, err := schema.TeamExists(db, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		modelExists, err := schema.ModelExists(db, modelId)
		if err != nil {
			return err
		}
		if !modelExists {
			return fmt.Errorf("model %v does not exists", modelId)
		}

		model, err := schema.GetModel(modelId, db, false, false, false)
		if err != nil {
			return err
		}

		if model.TeamId != nil {
			return fmt.Errorf("model %v is already assigned to team", modelId)
		}

		model.TeamId = &teamId

		result := db.Save(&model)
		if result.Error != nil {
			return schema.NewDbError("updating model team permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding model to team: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *TeamRouter) RemoveModelFromTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("model_id") {
		http.Error(w, "'team_id' and 'model_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, modelId := params.Get("team_id"), params.Get("model_id")

	err := t.db.Transaction(func(db *gorm.DB) error {
		teamExists, err := schema.TeamExists(db, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		modelExists, err := schema.ModelExists(db, modelId)
		if err != nil {
			return err
		}
		if !modelExists {
			return fmt.Errorf("model %v does not exists", modelId)
		}

		result := db.Model(&schema.Model{}).Where("id = ? and team_id = ?", modelId, teamId).Update("team_id", nil)
		if result.Error != nil {
			return schema.NewDbError("updating model team permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing model from team: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *TeamRouter) AddTeamAdmin(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	err := t.db.Transaction(func(db *gorm.DB) error {
		teamExists, err := schema.TeamExists(db, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(db, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := db.Save(&schema.UserTeam{TeamId: teamId, UserId: userId, IsTeamAdmin: true})
		if result.Error != nil {
			return schema.NewDbError("updating user team admin permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding team admin: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *TeamRouter) RemoveTeamAdmin(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	err := t.db.Transaction(func(db *gorm.DB) error {
		teamExists, err := schema.TeamExists(db, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(db, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := db.Save(&schema.UserTeam{TeamId: teamId, UserId: userId, IsTeamAdmin: false})
		if result.Error != nil {
			return schema.NewDbError("updating user team admin permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing team admin: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type TeamInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (t *TeamRouter) List(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, t.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var teams []schema.Team
	var result *gorm.DB
	if user.IsAdmin {
		result = t.db.Find(&teams)
	} else {
		userTeams := make([]string, 0, len(user.Teams))
		for _, t := range user.Teams {
			userTeams = append(userTeams, t.TeamId)
		}
		result = t.db.Where("id IN ?", userTeams).Find(&teams)
	}

	if result.Error != nil {
		err := schema.NewDbError("retrieving accessible teams", result.Error)
		http.Error(w, fmt.Sprintf("error listing teams: %v", err), http.StatusBadRequest)
		return
	}

	infos := make([]TeamInfo, 0, len(teams))
	for _, team := range teams {
		infos = append(infos, TeamInfo{Id: team.Id, Name: team.Name})
	}

	writeJsonResponse(w, infos)
}

type TeamUserInfo struct {
	UserId    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TeamAdmin bool   `json:"team_admin"`
}

func (t *TeamRouter) TeamUsers(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") {
		http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
		return
	}
	teamId := params.Get("team_id")

	teamExists, err := schema.TeamExists(t.db, teamId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !teamExists {
		http.Error(w, fmt.Sprintf("team %v does not exist", teamId), http.StatusBadRequest)
		return
	}

	var users []schema.UserTeam
	result := t.db.Preload("User").Where("team_id = ?", teamId).Find(&users)
	if result.Error != nil {
		err := schema.NewDbError("retrieving team users", result.Error)
		http.Error(w, fmt.Sprintf("error listing team users: %v", err), http.StatusBadRequest)
		return
	}

	infos := make([]TeamUserInfo, 0, len(users))
	for _, user := range users {
		infos = append(infos, TeamUserInfo{
			UserId:    user.UserId,
			Username:  user.User.Username,
			Email:     user.User.Email,
			TeamAdmin: user.IsTeamAdmin,
		})
	}

	writeJsonResponse(w, infos)
}

func (t *TeamRouter) TeamModels(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") {
		http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
		return
	}
	teamId := params.Get("team_id")

	teamExists, err := schema.TeamExists(t.db, teamId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !teamExists {
		http.Error(w, fmt.Sprintf("team %v does not exist", teamId), http.StatusBadRequest)
		return
	}

	var models []schema.Model
	result := t.db.
		Preload("Dependencies").Preload("Attributes").Preload("User").
		Or("access = ? AND team_id = ?", schema.Protected, teamId).
		Find(&models)

	if result.Error != nil {
		err := schema.NewDbError("retrieving team models", result.Error)
		http.Error(w, fmt.Sprintf("error listing team models: %v", err), http.StatusBadRequest)
		return
	}

	infos := make([]ModelInfo, 0, len(models))
	for _, model := range models {
		info, err := convertToModelInfo(model, t.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		infos = append(infos, info)
	}

	writeJsonResponse(w, infos)
}
