package auth

import (
	"fmt"
	"net/http"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type BasicIdentityProvider struct {
	jwtManager *JwtManager
	db         *gorm.DB
}

func NewBasicIdentityProvider(db *gorm.DB) IdentityProvider {
	return &BasicIdentityProvider{
		jwtManager: NewJwtManager(),
		db:         db,
	}
}

func (auth *BasicIdentityProvider) AuthMiddleware() chi.Middlewares {
	return chi.Middlewares{auth.jwtManager.Verifier(), auth.jwtManager.Authenticator()}
}

func (auth *BasicIdentityProvider) AllowDirectSignup() bool {
	return true
}

func (auth *BasicIdentityProvider) LoginWithEmail(email, password string) (LoginResult, error) {
	var user schema.User
	result := auth.db.Find(&user, "email = ?", email)
	if result.Error != nil {
		return LoginResult{}, schema.NewDbError("locating user for email", result.Error)
	}

	if result.RowsAffected != 1 {
		return LoginResult{}, fmt.Errorf("no user found with email %v", email)
	}

	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return LoginResult{}, fmt.Errorf("email and password do not match")
	}

	token, err := auth.jwtManager.CreateUserJwt(user.Id)
	if err != nil {
		return LoginResult{}, fmt.Errorf("login failed: %w", err)
	}

	return LoginResult{UserId: user.Id, AccessToken: token}, nil
}

func (auth *BasicIdentityProvider) LoginWithToken(accessToken string) (LoginResult, error) {
	return LoginResult{}, fmt.Errorf("login with token is not supported for this identity provider")
}

func (auth *BasicIdentityProvider) CreateUser(username, email, password string, admin bool) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", fmt.Errorf("error encrypting password: %w", err)
	}

	newUser := schema.User{Id: uuid.New().String(), Username: username, Email: email, Password: hashedPwd, IsAdmin: admin}

	err = auth.db.Transaction(func(txn *gorm.DB) error {
		var existingUser schema.User
		result := txn.Find(&existingUser, "username = ? or email = ?", username, email)
		if result.Error != nil {
			return schema.NewDbError("checking for existing username/email", result.Error)
		}
		if result.RowsAffected != 0 {
			if existingUser.Username == username {
				return ErrUsernameAlreadyExists
			} else {
				return ErrUserEmailAlreadyExists
			}
		}

		result = txn.Create(&newUser)
		if result.Error != nil {
			return schema.NewDbError("creating new user entry", result.Error)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error creating new user: %w", err)
	}

	return newUser.Id, nil
}

func (auth *BasicIdentityProvider) VerifyUser(userId string) error {
	return nil
}

func (auth *BasicIdentityProvider) DeleteUser(userId string) error {
	return nil
}

func (auth *BasicIdentityProvider) GetTokenExpiration(r *http.Request) (time.Time, error) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return time.Time{}, fmt.Errorf("error retrieving access token: %w", err)
	}

	return token.Expiration(), nil
}
