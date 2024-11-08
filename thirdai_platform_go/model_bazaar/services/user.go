package services

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db       *gorm.DB
	userAuth *auth.JwtManager
}

func (s *UserService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/signup", s.Signup)
		r.Post("/login", s.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())

		r.Get("/list", s.List)
		r.Get("/info", s.Info)
	})

	r.Group(func(r chi.Router) {
		r.Use(s.userAuth.Verifier())
		r.Use(s.userAuth.Authenticator())
		r.Use(auth.AdminOnly(s.db))

		r.Post("/promote-admin", s.PromoteAdmin)
		r.Post("/demote-admin", s.DemoteAdmin)
	})

	return r
}

type signupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signupResponse struct {
	UserId string `json:"user_id"`
}

func (s *UserService) CreateUser(username, email, password string, admin bool) (schema.User, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return schema.User{}, fmt.Errorf("error encrypting password: %w", err)
	}

	newUser := schema.User{Id: uuid.New().String(), Username: username, Email: email, Password: hashedPwd, IsAdmin: admin}

	err = s.db.Transaction(func(txn *gorm.DB) error {
		var existingUser schema.User
		result := txn.Find(&existingUser, "username = ? or email = ?", username, email)
		if result.Error != nil {
			return schema.NewDbError("checking for existing username/email", result.Error)
		}
		if result.RowsAffected != 0 {
			if existingUser.Username == username {
				return fmt.Errorf("username %v is already in use", username)
			} else {
				return fmt.Errorf("email %v is already in use", email)
			}
		}

		result = txn.Create(&newUser)
		if result.Error != nil {
			return schema.NewDbError("creating new user entry", result.Error)
		}

		return nil
	})

	if err != nil {
		return schema.User{}, fmt.Errorf("error creating new user: %w", err)
	}

	return newUser, nil
}

func (s *UserService) Signup(w http.ResponseWriter, r *http.Request) {
	var params signupRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	newUser, err := s.CreateUser(params.Username, params.Email, params.Password, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := signupResponse{UserId: newUser.Id}
	writeJsonResponse(w, res)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	UserId      string `json:"user_id"`
	AccessToken string `json:"access_token"`
}

func (s *UserService) Login(w http.ResponseWriter, r *http.Request) {
	var params loginRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	var user schema.User
	result := s.db.Find(&user, "email = ?", params.Email)
	if result.Error != nil {
		err := schema.NewDbError("locating user for email", result.Error)
		http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusBadRequest)
		return
	}
	if result.RowsAffected != 1 {
		http.Error(w, fmt.Sprintf("no user found with email %v", params.Email), http.StatusBadRequest)
		return
	}

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(params.Password))
	if err != nil {
		http.Error(w, "email and password do not match", http.StatusUnauthorized)
		return
	}

	token, err := s.userAuth.CreateUserJwt(user.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusInternalServerError)
		return
	}
	res := loginResponse{UserId: user.Id, AccessToken: token}
	writeJsonResponse(w, res)
}

type adminRequest struct {
	UserId string `json:"user_id"`
}

func (s *UserService) PromoteAdmin(w http.ResponseWriter, r *http.Request) {
	var params adminRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		user, err := schema.GetUser(params.UserId, txn, false)
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

	w.WriteHeader(http.StatusOK)
}

func (s *UserService) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
	var params adminRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	err := s.db.Transaction(func(txn *gorm.DB) error {
		user, err := schema.GetUser(params.UserId, txn, false)
		if err != nil {
			return err
		}

		var count int64
		result := txn.Model(&schema.User{}).Where("is_admin = ?", true).Count(&count)
		if result.Error != nil {
			return schema.NewDbError("counting existing admins", result.Error)
		}

		if count < 2 {
			return fmt.Errorf("cannot demote admin %v since there would be no admins left", params.UserId)
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

	w.WriteHeader(http.StatusOK)
}

type UserTeamInfo struct {
	TeamId    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	TeamAdmin bool   `json:"team_admin"`
}

type UserInfo struct {
	Id       string         `json:"id"`
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
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := schema.GetUser(userId, s.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var users []schema.User
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Preload("Teams").Preload("Teams.Team").Find(&users)
	} else if len(user.Teams) > 0 {
		userTeams := make([]string, 0, len(user.Teams))
		for _, t := range user.Teams {
			userTeams = append(userTeams, t.TeamId)
		}
		result = s.db.Preload("Teams").Preload("Teams.Team").Joins("JOIN user_teams ON user_teams.user_id = users.id").Where("user_teams.team_id in ?", userTeams).Find(&users)
	} else {
		users = []schema.User{user}
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
	writeJsonResponse(w, infos)
}

func (s *UserService) Info(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, s.db, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving user info: %v", err), http.StatusBadRequest)
	}

	info := convertToUserInfo(&user)
	writeJsonResponse(w, info)
}
