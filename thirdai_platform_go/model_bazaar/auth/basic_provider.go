package auth

import (
	"context"
	"errors"
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
	auditLog   AuditLogger
}

type BasicProviderArgs struct {
	Secret        []byte
	AdminUsername string
	AdminEmail    string
	AdminPassword string
}

func NewBasicIdentityProvider(db *gorm.DB, auditLog AuditLogger, args BasicProviderArgs) (IdentityProvider, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(args.AdminPassword), 10)
	if err != nil {
		return nil, fmt.Errorf("error encrypting admin password: %w", err)
	}

	err = addInitialAdminToDb(db, uuid.New().String(), args.AdminUsername, args.AdminEmail, hashedPwd)
	if err != nil {
		return nil, fmt.Errorf("error adding inital admin to db: %w", err)
	}

	return &BasicIdentityProvider{
		jwtManager: NewJwtManager(args.Secret),
		db:         db,
		auditLog:   auditLog,
	}, nil
}

func (auth *BasicIdentityProvider) addUserToContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			userId, err := ValueFromContext(r, userIdKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			user, err := schema.GetUser(userId, auth.db)
			if err != nil {
				http.Error(w, fmt.Sprintf("unable to find user %v: %v", userId, err), http.StatusUnauthorized)
				return
			}

			reqCtx := r.Context()
			reqCtx = context.WithValue(reqCtx, "user", user)
			next.ServeHTTP(w, r.WithContext(reqCtx))
		}

		return http.HandlerFunc(handler)
	}
}

func (auth *BasicIdentityProvider) AuthMiddleware() chi.Middlewares {
	return chi.Middlewares{auth.jwtManager.Verifier(), auth.jwtManager.Authenticator(), auth.addUserToContext(), auth.auditLog.Middleware}
}

func (auth *BasicIdentityProvider) AllowDirectSignup() bool {
	return true
}

func (auth *BasicIdentityProvider) LoginWithEmail(email, password string) (LoginResult, error) {
	var user schema.User
	result := auth.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return LoginResult{}, fmt.Errorf("no user found with email %v", email)
		}
		return LoginResult{}, schema.NewDbError("locating user for email", result.Error)
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

func (auth *BasicIdentityProvider) CreateUser(username, email, password string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", fmt.Errorf("error encrypting password: %w", err)
	}

	newUser := schema.User{Id: uuid.New().String(), Username: username, Email: email, Password: hashedPwd, IsAdmin: false}

	err = auth.db.Transaction(func(txn *gorm.DB) error {
		var existingUser schema.User
		result := txn.Limit(1).Find(&existingUser, "username = ? or email = ?", username, email)
		if result.Error != nil {
			return schema.NewDbError("checking for existing username/email", result.Error)
		}
		if result.RowsAffected != 0 {
			if existingUser.Username == username {
				return fmt.Errorf("username '%v' is already in use", username)
			} else {
				return fmt.Errorf("email '%v' is already in use", email)
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
