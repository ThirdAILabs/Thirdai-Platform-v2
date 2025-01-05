package auth

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
			return schema.NewDbError("checking if admin has already been added", result.Error)
		}
		if result.RowsAffected == 0 {
			result := txn.Create(&user)
			if result.Error != nil {
				return schema.NewDbError("creating initial admin", result.Error)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error adding initial admin to db: %w", err)
	}

	return nil
}
