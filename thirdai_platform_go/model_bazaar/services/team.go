package services

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/model_bazaar/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamService struct {
	db       *gorm.DB
	userAuth auth.IdentityProvider
}

func (s *TeamService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(s.userAuth.AuthMiddleware()...)

	r.With(auth.AdminOnly(s.db)).Post("/create", s.CreateTeam)

	r.Get("/list", s.List)

	r.Route("/{team_id}", func(r chi.Router) {
		r.With(auth.AdminOnly(s.db)).Delete("/", s.DeleteTeam)

		r.Group(func(r chi.Router) {
			r.Use(auth.AdminOrTeamAdminOnly(s.db))

			r.Post("/users/{user_id}", s.AddUserToTeam)
			r.Delete("/users/{user_id}", s.RemoveUserFromTeam)

			r.Post("/admins/{user_id}", s.AddTeamAdmin)
			r.Delete("/admins/{user_id}", s.RemoveTeamAdmin)

			r.Get("/users", s.TeamUsers)
			r.Get("/models", s.TeamModels)
		})

		r.Route("/models/{model_id}", func(r chi.Router) {
			r.Use(auth.TeamMemberOnly(s.db))
			r.Use(auth.ModelPermissionOnly(s.db, auth.OwnerPermission))

			r.Post("/", s.AddModelToTeam)
			r.Delete("/", s.RemoveModelFromTeam)
		})
	})

	return r
}

type createTeamRequest struct {
	Name string `json:"name"`
}

func (s *TeamService) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var params createTeamRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if params.Name == "" {
		http.Error(w, "Team name must be specified", http.StatusBadRequest)
	}

	newTeam := schema.Team{Id: uuid.New().String(), Name: params.Name}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		var existingTeam schema.Team
		result := txn.Find(&existingTeam, "name = ?", params.Name)
		if result.Error != nil {
			return schema.NewDbError("checking for existing team with name", result.Error)
		}
		if result.RowsAffected != 0 {
			return fmt.Errorf("team with name %v already exists", params.Name)
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

	utils.WriteJsonResponse(w, map[string]string{"team_id": newTeam.Id})
}

func (s *TeamService) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "team_id")

	team := schema.Team{Id: teamId}

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

	utils.WriteSuccess(w)
}

func (s *TeamService) AddUserToTeam(w http.ResponseWriter, r *http.Request) {
	teamId, userId := chi.URLParam(r, "team_id"), chi.URLParam(r, "user_id")

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

	utils.WriteSuccess(w)
}

func (s *TeamService) RemoveUserFromTeam(w http.ResponseWriter, r *http.Request) {
	teamId, userId := chi.URLParam(r, "team_id"), chi.URLParam(r, "user_id")

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

	utils.WriteSuccess(w)
}

func (s *TeamService) AddModelToTeam(w http.ResponseWriter, r *http.Request) {
	teamId, modelId := chi.URLParam(r, "team_id"), chi.URLParam(r, "model_id")

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

	utils.WriteSuccess(w)
}

func (s *TeamService) RemoveModelFromTeam(w http.ResponseWriter, r *http.Request) {
	teamId, modelId := chi.URLParam(r, "team_id"), chi.URLParam(r, "model_id")

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

	utils.WriteSuccess(w)
}

func (s *TeamService) AddTeamAdmin(w http.ResponseWriter, r *http.Request) {
	teamId, userId := chi.URLParam(r, "team_id"), chi.URLParam(r, "user_id")

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

	utils.WriteSuccess(w)
}

func (s *TeamService) RemoveTeamAdmin(w http.ResponseWriter, r *http.Request) {
	teamId, userId := chi.URLParam(r, "team_id"), chi.URLParam(r, "user_id")

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

	utils.WriteSuccess(w)
}

type TeamInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (s *TeamService) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var teams []schema.Team
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Find(&teams)
	} else {
		userTeams, err := schema.GetUserTeamIds(user.Id, s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
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

	utils.WriteJsonResponse(w, infos)
}

type TeamUserInfo struct {
	UserId    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TeamAdmin bool   `json:"team_admin"`
}

func (s *TeamService) TeamUsers(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "team_id")

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

	utils.WriteJsonResponse(w, infos)
}

func (s *TeamService) TeamModels(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "team_id")

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

	utils.WriteJsonResponse(w, infos)
}
