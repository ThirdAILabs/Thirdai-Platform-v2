package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFoundWithEmail = errors.New("no user found for given email")
	ErrInvalidCredentials    = errors.New("invalid login credentials")
	ErrGeneratingJwt         = errors.New("error generating jwt")
	ErrEmailAlreadyInUse     = errors.New("email is already in use")
	ErrUsernameAlreadyInUse  = errors.New("username is already in use")
)

type LoginResult struct {
	UserId      uuid.UUID
	AccessToken string
}

type IdentityProvider interface {
	AuthMiddleware() chi.Middlewares

	AllowDirectSignup() bool

	LoginWithEmail(email, password string) (LoginResult, error)

	LoginWithToken(accessToken string) (LoginResult, error)

	CreateUser(username, email, password string) (uuid.UUID, error)

	VerifyUser(userId uuid.UUID) error

	DeleteUser(userId uuid.UUID) error

	GetTokenExpiration(r *http.Request) (time.Time, error)
}

func addInitialAdminToDb(db *gorm.DB, userId uuid.UUID, username, email string, password []byte) error {
	user := schema.User{
		Id:       userId,
		Username: username,
		Email:    email,
		IsAdmin:  true,
	}
	if password != nil {
		user.Password = password
	}

	err := db.Transaction(func(txn *gorm.DB) error {
		var existingUser schema.User
		result := txn.Limit(1).Find(&existingUser, "id = ? or username = ? or email = ?", userId, username, email)
		if result.Error != nil {
			slog.Error("sql error checking if admin has already been added", "error", result.Error)
			return schema.ErrDbAccessFailed
		}
		if result.RowsAffected == 0 {
			result := txn.Create(&user)
			if result.Error != nil {
				slog.Error("sql error creating initial admin user", "error", result.Error)
				return schema.ErrDbAccessFailed
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error adding initial admin to db: %w", err)
	}

	return nil
}

type requestContextKey string

const (
	UserRequestContextKey requestContextKey = "user"
	ContextAPIKeyExpiry   requestContextKey = "api_key_expiry"
)
