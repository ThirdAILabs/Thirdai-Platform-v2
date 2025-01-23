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

type UserService struct {
	db       *gorm.DB
	userAuth auth.IdentityProvider
}

func (s *UserService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/signup", s.Signup)

		r.Get("/login", s.LoginWithEmail)
		r.Post("/login-with-token", s.LoginWithToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.AuthMiddleware()...)

		r.Get("/list", s.List)
		r.Get("/info", s.Info)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.AuthMiddleware()...)
		r.Use(auth.AdminOnly(s.db))

		r.Post("/create", s.CreateUser)

		r.Delete("/{user_id}", s.DeleteUser)

		r.Post("/{user_id}/admin", s.PromoteAdmin)
		r.Delete("/{user_id}/admin", s.DemoteAdmin)

		r.Post("/{user_id}/verify", s.VerifyUser)
	})

	return r
}

type signupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signupResponse struct {
	UserId uuid.UUID `json:"user_id"`
}

func (s *UserService) Signup(w http.ResponseWriter, r *http.Request) {
	var params signupRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	if !s.userAuth.AllowDirectSignup() {
		http.Error(w, "direct signup is not supported for this identify provider", http.StatusUnauthorized)
		return
	}

	userId, err := s.userAuth.CreateUser(params.Username, params.Email, params.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := signupResponse{UserId: userId}
	utils.WriteJsonResponse(w, res)
}

type loginResponse struct {
	UserId      uuid.UUID `json:"user_id"`
	AccessToken string    `json:"access_token"`
}

func (s *UserService) LoginWithEmail(w http.ResponseWriter, r *http.Request) {
	slog.Info("Got to login")

	email, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}

	login, err := s.userAuth.LoginWithEmail(email, password)
	if err != nil {
		http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusUnauthorized)
		return
	}

	res := loginResponse{UserId: login.UserId, AccessToken: login.AccessToken}
	utils.WriteJsonResponse(w, res)
}

type loginWithTokenRequest struct {
	AccessToken string `json:"access_token"`
}

func (s *UserService) LoginWithToken(w http.ResponseWriter, r *http.Request) {
	var params loginWithTokenRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	login, err := s.userAuth.LoginWithToken(params.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusUnauthorized)
		return
	}

	res := loginResponse{UserId: login.UserId, AccessToken: login.AccessToken}
	utils.WriteJsonResponse(w, res)
}

func (s *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		var admin schema.User
		adminResult := txn.Where("is_admin = ?", true).First(&admin)
		if adminResult.Error != nil {
			return schema.NewDbError("finding admin user", adminResult.Error)
		}

		updateResult := txn.Model(&schema.Model{}).
			Where("user_id = ?", userId).
			Update("user_id", admin.Id)
		if updateResult.Error != nil {
			return schema.NewDbError("updating owner of user protected/public models", updateResult.Error)
		}

		deleteUserResult := txn.Delete(&schema.User{Id: userId})
		if deleteUserResult.Error != nil {
			return schema.NewDbError("deleting user", deleteUserResult.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting user %v: %v", userId, err), http.StatusBadRequest)
		return
	}

	err = s.userAuth.DeleteUser(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting user %v: %v", userId, err), http.StatusBadRequest)
		return
	}

	utils.WriteSuccess(w)
}

func (s *UserService) PromoteAdmin(w http.ResponseWriter, r *http.Request) {
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		user, err := schema.GetUser(userId, txn)
		if err != nil {
			return err
		}

		user.IsAdmin = true

		result := txn.Save(&user)
		if result.Error != nil {
			return schema.NewDbError("updating user role to admin", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error promoting admin: %v", err), http.StatusBadRequest)
		return
	}

	utils.WriteSuccess(w)
}

func (s *UserService) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		user, err := schema.GetUser(userId, txn)
		if err != nil {
			return err
		}

		var count int64
		result := txn.Model(&schema.User{}).Where("is_admin = ?", true).Count(&count)
		if result.Error != nil {
			return schema.NewDbError("counting existing admins", result.Error)
		}

		if count < 2 {
			return fmt.Errorf("cannot demote admin %v since there would be no admins left", userId)
		}

		user.IsAdmin = false

		result = txn.Save(&user)
		if result.Error != nil {
			return schema.NewDbError("update user role to user", result.Error)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error demoting admin: %v", err), http.StatusBadRequest)
		return
	}

	utils.WriteSuccess(w)
}

type UserTeamInfo struct {
	TeamId    uuid.UUID `json:"team_id"`
	TeamName  string    `json:"team_name"`
	TeamAdmin bool      `json:"team_admin"`
}

type UserInfo struct {
	Id       uuid.UUID      `json:"id"`
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Admin    bool           `json:"admin"`
	Teams    []UserTeamInfo `json:"teams"`
}

func convertToUserInfo(user *schema.User) UserInfo {
	teams := make([]UserTeamInfo, 0, len(user.Teams))
	for _, team := range user.Teams {
		teams = append(teams, UserTeamInfo{
			TeamId:    team.TeamId,
			TeamName:  team.Team.Name,
			TeamAdmin: team.IsTeamAdmin,
		})
	}

	return UserInfo{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
		Admin:    user.IsAdmin,
		Teams:    teams,
	}
}

func (s *UserService) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var users []schema.User
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Preload("Teams").Preload("Teams.Team").Find(&users)
	} else {
		userTeams, err := schema.GetUserTeamIds(user.Id, s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(userTeams) > 0 {
			result = s.db.Preload("Teams").Preload("Teams.Team").Joins("JOIN user_teams ON user_teams.user_id = users.id").Where("user_teams.team_id in ?", userTeams).Find(&users)
		} else {
			users = []schema.User{user}
		}
	}

	if result != nil && result.Error != nil {
		err := schema.NewDbError("retrieving list of users", result.Error)
		http.Error(w, fmt.Sprintf("error listing users: %v", err), http.StatusBadRequest)
		return
	}

	infos := make([]UserInfo, 0, len(users))
	for _, u := range users {
		infos = append(infos, convertToUserInfo(&u))
	}
	utils.WriteJsonResponse(w, infos)
}

func (s *UserService) Info(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userWithTeams schema.User
	result := s.db.Preload("Teams").Preload("Teams.Team").First(&userWithTeams, "id = ?", user.Id)
	if result.Error != nil {
		err := schema.NewDbError("retrieving user info with teams", result.Error)
		http.Error(w, fmt.Sprintf("error getting user info: %v", err), http.StatusBadRequest)
		return
	}

	info := convertToUserInfo(&userWithTeams)
	utils.WriteJsonResponse(w, info)
}

func (s *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params signupRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	userId, err := s.userAuth.CreateUser(params.Username, params.Email, params.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusBadRequest)
		return
	}

	res := signupResponse{UserId: userId}
	utils.WriteJsonResponse(w, res)
}

func (s *UserService) VerifyUser(w http.ResponseWriter, r *http.Request) {
	userId, err := utils.URLParamUUID(r, "user_id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.userAuth.VerifyUser(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error verifying user '%v': %v", userId, err), http.StatusBadRequest)
		return
	}

	utils.WriteSuccess(w)
}
