package routers

import (
	"fmt"
	"net/http"
	"thirdai_platform/src/auth"
	"thirdai_platform/src/schema"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRouter struct {
	db           *gorm.DB
	tokenManager *auth.JwtManager
}

func (u *UserRouter) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/signup", u.Signup)
		r.Post("/login", u.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(u.tokenManager.Verifier())
		r.Use(u.tokenManager.Authenticator())

		r.Get("/list", u.List)
		r.Get("/info", u.Info)
	})

	r.Group(func(r chi.Router) {
		r.Use(u.tokenManager.Verifier())
		r.Use(u.tokenManager.Authenticator())
		r.Use(u.tokenManager.AdminOnly(u.db))

		r.Post("/promote-admin", u.PromoteAdmin)
		r.Post("/demote-admin", u.DemoteAdmin)
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

func (u *UserRouter) Signup(w http.ResponseWriter, r *http.Request) {
	var params signupRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	var existingUser schema.User
	result := u.db.Find(&existingUser, "username = ? or email = ?", params.Username, params.Email)
	if result.Error != nil {
		dbError(w, result.Error)
		return
	}
	if result.RowsAffected != 0 {
		if existingUser.Username == params.Username {
			http.Error(w, fmt.Sprintf("username %v is already in use", params.Username), http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("email %v is already in use", params.Email), http.StatusBadRequest)
		}
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	if err != nil {
		http.Error(w, "error encrypting password", http.StatusInternalServerError)
		return
	}

	newUser := schema.User{Id: uuid.New().String(), Username: params.Username, Email: params.Email, Password: hashedPwd}

	// TODO(Nicholas): should this be done in a transaction to handle edge case of
	// concurrent signups? Not correctness issue, just for better error messages to user
	result = u.db.Create(&newUser)
	if result.Error != nil {
		dbError(w, result.Error)
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
	AccessToken string `json:"access_token"`
}

func (u *UserRouter) Login(w http.ResponseWriter, r *http.Request) {
	var params loginRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	var user schema.User
	result := u.db.Find(&user, "email = ?", params.Email)
	if result.Error != nil {
		dbError(w, result.Error)
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

	token, err := u.tokenManager.CreateToken(user.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating access token: %v", err), http.StatusInternalServerError)
		return
	}
	res := loginResponse{AccessToken: token}
	writeJsonResponse(w, res)
}

type adminRequest struct {
	UserId string `json:"user_id"`
}

func (u *UserRouter) PromoteAdmin(w http.ResponseWriter, r *http.Request) {
	var params adminRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	user, err := schema.GetUser(params.UserId, u.db, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.IsAdmin = true

	result := u.db.Save(&user)
	if result.Error != nil {
		dbError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (u *UserRouter) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
	var params adminRequest
	if !parseRequestBody(w, r, &params) {
		return
	}

	user, err := schema.GetUser(params.UserId, u.db, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = u.db.Transaction(func(db *gorm.DB) error {
		var count int64
		result := db.Model(&schema.User{}).Where("is_admin = ?", true).Count(&count)
		if result.Error != nil {
			return fmt.Errorf("database error: %v", err)
		}

		if count < 2 {
			return fmt.Errorf("cannot demote admin %v since there would be no admins left", params.UserId)
		}

		user.IsAdmin = false

		result = u.db.Save(&user)
		if result.Error != nil {
			return fmt.Errorf("database error: %v", err)
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type teamInfo struct {
	TeamId    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	TeamAdmin bool   `json:"team_admin"`
}

type userInfo struct {
	Id       string     `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Admin    bool       `json:"admin"`
	Teams    []teamInfo `json:"teams"`
}

func convertToUserInfo(user *schema.User) userInfo {
	teams := make([]teamInfo, 0, len(user.Teams))
	for _, team := range user.Teams {
		teams = append(teams, teamInfo{
			TeamId:    team.TeamId,
			TeamName:  team.Team.Name,
			TeamAdmin: team.IsTeamAdmin,
		})
	}

	return userInfo{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
		Admin:    user.IsAdmin,
		Teams:    teams,
	}
}

func (u *UserRouter) List(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, u.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var users []schema.User
	var result *gorm.DB
	if user.IsAdmin {
		result = u.db.Preload("Teams").Preload("Teams.Team").Find(&users)
	} else if len(user.Teams) > 0 {
		userTeams := make([]string, 0, len(user.Teams))
		for _, t := range user.Teams {
			userTeams = append(userTeams, t.TeamId)
		}
		result = u.db.Preload("Teams").Preload("Teams.Team").Joins("JOIN user_teams ON user_teams.user_id = users.id").Where("user_teams.team_id in ?", userTeams).Find(&users)
	} else {
		users = []schema.User{user}
	}

	if result != nil && result.Error != nil {
		dbError(w, result.Error)
		return
	}

	infos := make([]userInfo, 0, len(users))
	for _, u := range users {
		infos = append(infos, convertToUserInfo(&u))
	}
	writeJsonResponse(w, infos)
}

func (u *UserRouter) Info(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.UserIdFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := schema.GetUser(userId, u.db, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	info := convertToUserInfo(&user)
	writeJsonResponse(w, info)
}
