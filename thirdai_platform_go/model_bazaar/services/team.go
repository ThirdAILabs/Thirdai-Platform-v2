package services

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamService struct {
	db       *gorm.DB
	userAuth *auth.JwtManager
}

func (s *TeamService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.AdminOnly(s.db))

		r.Post("/create", s.CreateTeam)
		r.Post("/delete", s.DeleteTeam)

	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.AdminOrTeamAdminOnly(s.db))

		r.Post("/add-user", s.AddUserToTeam)
		r.Post("/remove-user", s.RemoveUserFromTeam)

		r.Post("/add-admin", s.AddTeamAdmin)
		r.Post("/remove-admin", s.RemoveTeamAdmin)

		r.Get("/users", s.TeamUsers)
		r.Get("/models", s.TeamModels)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.ModelPermissionOnly(s.db, auth.OwnerPermission))

		r.Post("/add-model", s.AddModelToTeam)
		r.Post("/remove-model", s.RemoveModelFromTeam)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())

		r.Get("/list", s.List)
	})

	return r
}

func (s *TeamService) CreateTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("name") {
		http.Error(w, "'name' query parameter missing", http.StatusBadRequest)
		return
	}
	name := params.Get("name")

	newTeam := schema.Team{Id: uuid.New().String(), Name: name}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		var existingTeam schema.Team
		result := txn.Find(&existingTeam, "name = ?", name)
		if result.Error != nil {
			return schema.NewDbError("checking for existing team with name", result.Error)
		}
		if result.RowsAffected != 0 {
			return fmt.Errorf("team with name %v already exists", name)
		}

		result = txn.Create(&newTeam)
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

func (s *TeamService) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") {
		http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
		return
	}

	team := schema.Team{Id: params.Get("team_id")}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		exists, err := schema.TeamExists(txn, team.Id)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("team %v does not exists", team.Id)
		}

		result := txn.Delete(&team)
		if result.Error != nil {
			return schema.NewDbError("deleting team", result.Error)
		}

		result = txn.Model(&schema.Model{}).Where("team_id = ?", team.Id).Update("team_id", nil).Update("access", schema.Private)
		if result.Error != nil {
			return schema.NewDbError("updating access for models after deleting team", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting team: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

func (s *TeamService) AddUserToTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	userTeam := schema.UserTeam{UserId: userId, TeamId: teamId}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		teamExists, err := schema.TeamExists(txn, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(txn, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := txn.Create(&userTeam)
		if result.Error != nil {
			return schema.NewDbError("creating new user_team entry", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding user to team: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

func (s *TeamService) RemoveUserFromTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
		teamExists, err := schema.TeamExists(txn, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(txn, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := txn.Delete(&schema.UserTeam{UserId: userId, TeamId: teamId})
		if result.Error != nil {
			return schema.NewDbError("deleting user_team entry", result.Error)
		}

		result = txn.Model(&schema.Model{}).Where("team_id = ? and user_id = ?", teamId, userId).Update("team_id", nil).Update("access", schema.Private)
		if result.Error != nil {
			return schema.NewDbError("updating model permissions after removing user from team", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing user from team: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

func (s *TeamService) AddModelToTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("model_id") {
		http.Error(w, "'team_id' and 'model_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, modelId := params.Get("team_id"), params.Get("model_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
		teamExists, err := schema.TeamExists(txn, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		modelExists, err := schema.ModelExists(txn, modelId)
		if err != nil {
			return err
		}
		if !modelExists {
			return fmt.Errorf("model %v does not exists", modelId)
		}

		model, err := schema.GetModel(modelId, txn, false, false, false)
		if err != nil {
			return err
		}

		if model.TeamId != nil {
			return fmt.Errorf("model %v is already assigned to team", modelId)
		}

		model.TeamId = &teamId

		result := txn.Save(&model)
		if result.Error != nil {
			return schema.NewDbError("updating model team permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding model to team: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

func (s *TeamService) RemoveModelFromTeam(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("model_id") {
		http.Error(w, "'team_id' and 'model_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, modelId := params.Get("team_id"), params.Get("model_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
		teamExists, err := schema.TeamExists(txn, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		modelExists, err := schema.ModelExists(txn, modelId)
		if err != nil {
			return err
		}
		if !modelExists {
			return fmt.Errorf("model %v does not exists", modelId)
		}

		result := txn.Model(&schema.Model{}).Where("id = ? and team_id = ?", modelId, teamId).Update("team_id", nil)
		if result.Error != nil {
			return schema.NewDbError("updating model team permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing model from team: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

func (s *TeamService) AddTeamAdmin(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
		teamExists, err := schema.TeamExists(txn, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(txn, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := txn.Save(&schema.UserTeam{TeamId: teamId, UserId: userId, IsTeamAdmin: true})
		if result.Error != nil {
			return schema.NewDbError("updating user team admin permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding team admin: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

func (s *TeamService) RemoveTeamAdmin(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") || !params.Has("user_id") {
		http.Error(w, "'team_id' and 'user_id' query parameters missing", http.StatusBadRequest)
		return
	}
	teamId, userId := params.Get("team_id"), params.Get("user_id")

	err := s.db.Transaction(func(txn *gorm.DB) error {
		teamExists, err := schema.TeamExists(txn, teamId)
		if err != nil {
			return err
		}
		if !teamExists {
			return fmt.Errorf("team %v does not exists", teamId)
		}

		userExists, err := schema.UserExists(txn, userId)
		if err != nil {
			return err
		}
		if !userExists {
			return fmt.Errorf("user %v does not exists", userId)
		}

		result := txn.Save(&schema.UserTeam{TeamId: teamId, UserId: userId, IsTeamAdmin: false})
		if result.Error != nil {
			return schema.NewDbError("updating user team admin permission", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing team admin: %v", err), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

type TeamInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (s *TeamService) List(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, s.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var teams []schema.Team
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Find(&teams)
	} else {
		userTeams := make([]string, 0, len(user.Teams))
		for _, t := range user.Teams {
			userTeams = append(userTeams, t.TeamId)
		}
		result = s.db.Where("id IN ?", userTeams).Find(&teams)
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

func (s *TeamService) TeamUsers(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") {
		http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
		return
	}
	teamId := params.Get("team_id")

	teamExists, err := schema.TeamExists(s.db, teamId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !teamExists {
		http.Error(w, fmt.Sprintf("team %v does not exist", teamId), http.StatusBadRequest)
		return
	}

	var users []schema.UserTeam
	result := s.db.Preload("User").Where("team_id = ?", teamId).Find(&users)
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

func (s *TeamService) TeamModels(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if !params.Has("team_id") {
		http.Error(w, "'team_id' query parameter missing", http.StatusBadRequest)
		return
	}
	teamId := params.Get("team_id")

	teamExists, err := schema.TeamExists(s.db, teamId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !teamExists {
		http.Error(w, fmt.Sprintf("team %v does not exist", teamId), http.StatusBadRequest)
		return
	}

	var models []schema.Model
	result := s.db.
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
		info, err := convertToModelInfo(model, s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		infos = append(infos, info)
	}

	writeJsonResponse(w, infos)
}
