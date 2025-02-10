package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/utils"

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
	})

	return r
}

type createTeamRequest struct {
	Name string `json:"name"`
}

type createTeamResponse struct {
	TeamId uuid.UUID `json:"team_id"`
}

func (s *TeamService) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var params createTeamRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if params.Name == "" {
		http.Error(w, "Team name must be specified", http.StatusBadRequest)
		return
	}

	newTeam := schema.Team{Id: uuid.New(), Name: params.Name}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		var existingTeam schema.Team
		result := txn.Limit(1).Find(&existingTeam, "name = ?", params.Name)
		if result.Error != nil {
			slog.Error("sql error checking for duplicate team name", "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}
		if result.RowsAffected != 0 {
			return CodedError(fmt.Errorf("team with name %v already exists", params.Name), http.StatusConflict)
		}

		result = txn.Create(&newTeam)
		if result.Error != nil {
			slog.Error("sql error creating new team", "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error creating team: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteJsonResponse(w, createTeamResponse{TeamId: newTeam.Id})
}

func (s *TeamService) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	team := schema.Team{Id: teamId}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkTeamExists(txn, team.Id); err != nil {
			return err
		}

		result := txn.Delete(&team)
		if result.Error != nil {
			slog.Error("sql error deleting team", "team_id", teamId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		result = txn.Model(&schema.Model{}).Where("team_id = ?", team.Id).Update("team_id", nil).Update("access", schema.Private)
		if result.Error != nil {
			slog.Error("sql error updating model permissions after team deletion", "team_id", teamId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting team: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

func (s *TeamService) AddUserToTeam(w http.ResponseWriter, r *http.Request) {
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userTeam := schema.UserTeam{UserId: userId, TeamId: teamId}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkTeamExists(txn, teamId); err != nil {
			return err
		}

		if err := checkUserExists(txn, userId); err != nil {
			return err
		}

		result := txn.Create(&userTeam)
		if result.Error != nil {
			slog.Error("sql error creating new user_team entry", "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding user to team: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

func (s *TeamService) RemoveUserFromTeam(w http.ResponseWriter, r *http.Request) {
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkTeamExists(txn, teamId); err != nil {
			return err
		}

		if err := checkUserExists(txn, userId); err != nil {
			return err
		}

		if err := checkTeamMember(txn, userId, teamId); err != nil {
			return err
		}

		result := txn.Delete(&schema.UserTeam{UserId: userId, TeamId: teamId})
		if result.Error != nil {
			slog.Error("sql error deleting user_team entry", "team_id", teamId, "user_id", userId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		result = txn.Model(&schema.Model{}).Where("team_id = ? and user_id = ?", teamId, userId).Update("team_id", nil).Update("access", schema.Private)
		if result.Error != nil {
			slog.Error("sql error updating model permissions after removing user from team", "team_id", teamId, "user_id", userId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing user from team: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

func (s *TeamService) AddTeamAdmin(w http.ResponseWriter, r *http.Request) {
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkTeamExists(txn, teamId); err != nil {
			return err
		}

		if err := checkUserExists(txn, userId); err != nil {
			return err
		}

		result := txn.Save(&schema.UserTeam{TeamId: teamId, UserId: userId, IsTeamAdmin: true})
		if result.Error != nil {
			slog.Error("sql error updating user to team admin", "user_id", userId, "team_id", teamId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error adding team admin: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

func (s *TeamService) RemoveTeamAdmin(w http.ResponseWriter, r *http.Request) {
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		if err := checkTeamExists(txn, teamId); err != nil {
			return err
		}

		if err := checkUserExists(txn, userId); err != nil {
			return err
		}

		if err := checkTeamMember(txn, userId, teamId); err != nil {
			return err
		}

		result := txn.Model(&schema.UserTeam{TeamId: teamId, UserId: userId}).Update("is_team_admin", false)
		if result.Error != nil {
			slog.Error("sql error removing user team admin permission", "user_id", userId, "team_id", teamId, "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error removing team admin: %v", err), GetResponseCode(err))
		return
	}

	utils.WriteSuccess(w)
}

type TeamInfo struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (s *TeamService) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var teams []schema.Team
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Find(&teams)
	} else {
		userTeams, err := schema.GetUserTeamIds(user.Id, s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = s.db.Where("id IN ?", userTeams).Find(&teams)
	}

	if result.Error != nil {
		slog.Error("sql error listing accessible teams", "user_id", user.Id, "error", result.Error)
		http.Error(w, fmt.Sprintf("error listing teams: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}

	infos := make([]TeamInfo, 0, len(teams))
	for _, team := range teams {
		infos = append(infos, TeamInfo{Id: team.Id, Name: team.Name})
	}

	utils.WriteJsonResponse(w, infos)
}

type TeamUserInfo struct {
	UserId    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	TeamAdmin bool      `json:"team_admin"`
}

func (s *TeamService) TeamUsers(w http.ResponseWriter, r *http.Request) {
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := checkTeamExists(s.db, teamId); err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	var users []schema.UserTeam
	result := s.db.Preload("User").Where("team_id = ?", teamId).Find(&users)
	if result.Error != nil {
		slog.Error("sql error listing team users", "team_id", teamId, "error", result.Error)
		http.Error(w, fmt.Sprintf("error listing team users: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
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
	teamId, err := utils.URLParamUUID(r, "team_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := checkTeamExists(s.db, teamId); err != nil {
		http.Error(w, err.Error(), GetResponseCode(err))
		return
	}

	var models []schema.Model
	result := s.db.
		Preload("Dependencies").Preload("Attributes").Preload("User").
		Or("access = ? AND team_id = ?", schema.Protected, teamId).
		Find(&models)

	if result.Error != nil {
		slog.Error("sql error listing team models", "team_id", teamId, "error", result.Error)
		http.Error(w, fmt.Sprintf("error listing team models: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}

	infos := make([]ModelInfo, 0, len(models))
	for _, model := range models {
		info, err := convertToModelInfo(model, s.db)
		if err != nil {
			http.Error(w, err.Error(), GetResponseCode(err))
			return
		}
		infos = append(infos, info)
	}

	utils.WriteJsonResponse(w, infos)
}
