package services

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/utils"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
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
		if s.userAuth.AllowDirectSignup() {
			r.Post("/signup", s.Signup)
		}

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
		http.Error(w, "direct signup is not supported for this identify provider", http.StatusBadRequest)
		return
	}

	userId, err := s.userAuth.CreateUser(params.Username, params.Email, params.Password)
	if err != nil {
		responseCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, auth.ErrEmailAlreadyInUse):
			responseCode = http.StatusConflict
		case errors.Is(err, auth.ErrUsernameAlreadyInUse):
			responseCode = http.StatusConflict
		}
		http.Error(w, err.Error(), responseCode)
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
		responseCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, auth.ErrUserNotFoundWithEmail):
			responseCode = http.StatusNotFound
		case errors.Is(err, auth.ErrInvalidCredentials):
			responseCode = http.StatusUnauthorized
		}
		http.Error(w, fmt.Sprintf("login failed: %v", err), responseCode)
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
		// This can only fail if keycloak cannot provide information about the user, which
		// should not happen if they are already signed in, thus this is a internal server error.
		// TODO(any): techically this could be http.StatusUnauthorized if an invalid token is
		// provided, however it's not clear how the client will report this error.
		http.Error(w, fmt.Sprintf("login failed: %v", err), http.StatusInternalServerError)
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
			slog.Error("sql error finding admin to assign models to", "user_id", userId, "error", adminResult.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		updateResult := txn.Model(&schema.Model{}).
			Where("user_id = ?", userId).
			Update("user_id", admin.Id)
		if updateResult.Error != nil {
			slog.Error("sql error updating owner of user protected/public models", "user_id", userId, "error", updateResult.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		deleteUserResult := txn.Delete(&schema.User{Id: userId})
		if deleteUserResult.Error != nil {
			slog.Error("sql error deleting user", "user_id", userId, "error", err)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting user %v: %v", userId, err), GetResponseCode(err))
		return
	}

	err = s.userAuth.DeleteUser(userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting user %v: %v", userId, err), http.StatusInternalServerError)
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
			if errors.Is(err, schema.ErrUserNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		user.IsAdmin = true

		result := txn.Save(&user)
		if result.Error != nil {
			slog.Error("sql error updating user role to admin", "user_id", userId, "error", err)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error promoting admin: %v", err), GetResponseCode(err))
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
			if errors.Is(err, schema.ErrUserNotFound) {
				return CodedError(err, http.StatusNotFound)
			}
			return CodedError(err, http.StatusInternalServerError)
		}

		if !user.IsAdmin {
			return CodedError(errors.New("user is already not an admim"), http.StatusUnprocessableEntity)
		}

		var count int64
		result := txn.Model(&schema.User{}).Where("is_admin = ?", true).Count(&count)
		if result.Error != nil {
			slog.Error("sql error counting existing admins", "error", result.Error)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		if count < 2 {
			return CodedError(fmt.Errorf("cannot demote admin %v since there would be no admins left", userId), http.StatusUnprocessableEntity)
		}

		user.IsAdmin = false

		result = txn.Save(&user)
		if result.Error != nil {
			slog.Error("sql error updating user role to user", "user_id", userId, "error", err)
			return CodedError(schema.ErrDbAccessFailed, http.StatusInternalServerError)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("error demoting admin: %v", err), GetResponseCode(err))
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
	Id            uuid.UUID      `json:"id"`
	Username      string         `json:"username"`
	Email         string         `json:"email"`
	Admin         bool           `json:"admin"`
	Teams         []UserTeamInfo `json:"teams"`
	RoleSignature string         `json:"role_signature"`
}

type RolePayload struct {
	Admin bool           `json:"global_admin"`
	Teams []UserTeamInfo `json:"teams"`
	jwt.RegisteredClaims
}

func SignRolePayload(payload RolePayload) (string, error) {
	var privateKey *rsa.PrivateKey
	privateKeyData := `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDgyAUl9RGfWdeX
hyo+ByzffCs7XaAf7+PDp3u9s6woYVe4wppauECkBy7+2i6/Iaqw95D+8mSHXDK7
akf8tDysNcfrsyzTaXl2ZS4k7XHMgCWdUs50r7eI///YT+Wbt3GQnPs23ZL+y6Kc
PGph+hletaXS3i7YfoSFA/N6tx+49wi1tmn83ZEnSUA778JDjc2XOfz3xHq2zLoK
HB05kKXVg/kuymktYU2MxyPTa4U+sHf7pxYqP26XnzCuv8Sl5bHdhaRsHCyfYxJ9
Uw13fmT9kYNaYDjNN2DvnVMAtKBkNX4Zm+nb2rfPbZZxe8ncuBvoqFJVwdQepbGQ
BSymxgCpAgMBAAECggEAAK9ShuSnbkQiiLKUdBfgjaXcU48z5EQ398IQTPoWigRB
tXCGjASQC6fkpa/7im3JwWw5rJf6t8f1209GAzmd2zT211hTO6nDoT/KhPm7ugCn
pkiieov8AktflIF2z/NrKBW2we2pWF3wturlgWFDxHkt27wwg2Yyrp1eCsRMapHF
pmWTtBNW1xiWQbmfJE5UASWYVQI5Q5rdpZLv2vXprtAA+uSE4JFtvezqkW98yjso
QmkqQNUVOjmOZ2jKtgIO6kAh+et9vnFl79XVozW1jrM1hFOSFzr86sp3z/LmJdyC
R/QMLXrwFJrKFwo2IHZKDP2P+Y3lA+Y7E08m+EMH4QKBgQD2jx07O3TfGhcGes5B
aVEBK0yvKBlwFBxHENC2lqWBI4UsdLej5OUODAYqF6YnW8Azq6sAxKk7WdzwZWQY
nsKYqmb43ftaqP5suJGqf/yw/96dYf8aczSHKJHEYZZmMeTZO1iM+QMFE2eNXpW2
R3Klvxzzpso+JR/RmpysNQOWiQKBgQDpY25DvxQfNTHF94TX1qd0V6uwNNEvtxFv
v+j+2tuz3xwSow+i1/mXSc5ZeIS/yCWEfVbOFZIFVS1+GFLlThuZyAlioiNuRM17
1WUkupyjBgsVfXGKC5MvMP8NfoGwbGJ5hKQOa0S5w3dYr4kSbl8ebf2l4b6rGtSH
7KeTPayRIQKBgQDeeU5YBxMyyHjkSOVZUm1cT7S3C8jAP/UwDrU1PAOE3gcpkPuv
MDeakDDzxDkRpJFuTkVTwSAuxKw+Yk6KhJ50YLXfc3V9XaWNdpFBtpDNKWO2wRkN
xcws9Odqut+ZwQWNGiaRtZMK/nJetm0Cd7+0XRkDpYkxwA/Q8uDR5lgheQKBgQC9
9b4jyfy4wfU3KpWnkAFQAqOtke/JpHm+uTcNaFl2d9xDlxD8/EkcSGh6DkwORPu0
cMgciRYG3SNgBLBED2ULr/NjopCwCbQuXKwsTu97CUowPaASOgWcXYbbFuK8FBu6
yKk3Szvu7xfOyWEJ7WfiPqg7QhiM8BOYZpimkYZJwQKBgQDqZvWjXm/ls/jj7BE4
EeYsKGDYsiTnkwGZfpxTD3vMgh8QCM0h4Ouq2gph6Q3dGHOvBAZ8lrUoGJYivP1d
jvSxHiMtyZAoH4EhTVAjlkAavOQu7vYC0U3QIwHJXme4H7bsSq11i9mrMhOgbJrD
wsrLbFp5MB71vcT+xNKuEFWuyQ==
-----END PRIVATE KEY-----`

	var err error
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyData))
	if err != nil {
		slog.Error("Failed to load RSA keys", "error", err.Error())
		return "", fmt.Errorf("error loading private key: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)
	return token.SignedString(privateKey)
}

func convertToUserInfo(user *schema.User) (UserInfo, error) {
	teams := make([]UserTeamInfo, 0, len(user.Teams))
	for _, team := range user.Teams {
		teams = append(teams, UserTeamInfo{
			TeamId:    team.TeamId,
			TeamName:  team.Team.Name,
			TeamAdmin: team.IsTeamAdmin,
		})
	}

	rolePayload := RolePayload{
		Admin: user.IsAdmin,
		Teams: teams,
	}

	roleSignature, err := SignRolePayload(rolePayload)
	if err != nil {
		slog.Error("Error signing role payload", "error", err.Error())
		return UserInfo{}, err
	}

	return UserInfo{
		Id:            user.Id,
		Username:      user.Username,
		Email:         user.Email,
		Admin:         user.IsAdmin,
		Teams:         teams,
		RoleSignature: roleSignature,
	}, nil
}

func (s *UserService) List(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var users []schema.User
	var result *gorm.DB
	if user.IsAdmin {
		result = s.db.Preload("Teams").Preload("Teams.Team").Find(&users)
	} else {
		userTeams, err := schema.GetUserTeamIds(user.Id, s.db)
		if err != nil {
			http.Error(w, "error loading user teams", http.StatusInternalServerError)
			return
		}
		if len(userTeams) > 0 {
			result = s.db.Preload("Teams").Preload("Teams.Team").Joins("JOIN user_teams ON user_teams.user_id = users.id").Where("user_teams.team_id in ?", userTeams).Find(&users)
		} else {
			users = []schema.User{user}
		}
	}

	if result != nil && result.Error != nil {
		slog.Error("sql error listing users", "error", result.Error)
		http.Error(w, fmt.Sprintf("error listing users: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}

	infos := make([]UserInfo, 0, len(users))
	for _, u := range users {
		info, err := convertToUserInfo(&u)
		if err != nil {
			slog.Error("error converting user to user info", "error", err)
			http.Error(w, fmt.Sprintf("error converting user to user info: %v", err), http.StatusInternalServerError)
		}
		infos = append(infos, info)
	}
	utils.WriteJsonResponse(w, infos)
}

func (s *UserService) Info(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userWithTeams schema.User
	result := s.db.Preload("Teams").Preload("Teams.Team").First(&userWithTeams, "id = ?", user.Id)
	if result.Error != nil {
		slog.Error("sql error loading user info", "user_id", user.Id, "error", result.Error)
		http.Error(w, fmt.Sprintf("error getting user info: %v", schema.ErrDbAccessFailed), http.StatusInternalServerError)
		return
	}

	info, err := convertToUserInfo(&userWithTeams)
	if err != nil {
		http.Error(w, fmt.Sprintf("error converting user to user info: %v", err), http.StatusInternalServerError)
		return
	}
	utils.WriteJsonResponse(w, info)
}

func (s *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params signupRequest
	if !utils.ParseRequestBody(w, r, &params) {
		return
	}

	userId, err := s.userAuth.CreateUser(params.Username, params.Email, params.Password)
	if err != nil {
		responseCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, auth.ErrEmailAlreadyInUse):
			responseCode = http.StatusConflict
		case errors.Is(err, auth.ErrUsernameAlreadyInUse):
			responseCode = http.StatusConflict
		}
		http.Error(w, fmt.Sprintf("error creating user: %v", err), responseCode)
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
		http.Error(w, fmt.Sprintf("error verifying user '%v': %v", userId, err), http.StatusInternalServerError)
		return
	}

	utils.WriteSuccess(w)
}
