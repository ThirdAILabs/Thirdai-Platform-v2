package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type LoginResult struct {
	UserId      string
	AccessToken string
}

type IdentityProvider interface {
	AuthMiddleware() chi.Middlewares

	AllowDirectSignup() bool

	LoginWithEmail(email, password string) (LoginResult, error)

	LoginWithToken(accessToken string) (LoginResult, error)

	CreateUser(username, email, password string, admin bool) (string, error)

	VerifyUser(userId string) error

	DeleteUser(userId string) error

	GetTokenExpiration(r *http.Request) (time.Time, error)
}

var ErrUserEmailAlreadyExists = errors.New("email is already in use")
var ErrUsernameAlreadyExists = errors.New("username is already in use")
