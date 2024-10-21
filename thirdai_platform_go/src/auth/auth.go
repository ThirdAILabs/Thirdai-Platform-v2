package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type JwtManager struct {
	auth *jwtauth.JWTAuth
}

func (m *JwtManager) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(m.auth)
}

func (m *JwtManager) Authenticator() func(http.Handler) http.Handler {
	return jwtauth.Authenticator(m.auth)
}

const userIdKey = "user_id"

func (m *JwtManager) CreateToken(user_id string) (string, error) {
	claims := map[string]interface{}{
		userIdKey: user_id,
		"exp":     time.Now().Add(time.Minute * 15),
	}
	_, token, err := m.auth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}
	return token, nil
}

func UserIdFromContext(r *http.Request) (string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", fmt.Errorf("error retrieving auth claims: %v", err)
	}

	userIdUncasted, ok := claims[userIdKey]
	if !ok {
		return "", fmt.Errorf("invalid token: unable to locate user_id")
	}

	userId, ok := userIdUncasted.(string)
	if !ok {
		return "", fmt.Errorf("invalid token: user_id is invalid type")
	}

	return userId, nil
}
