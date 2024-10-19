package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type Manager struct {
	auth *jwtauth.JWTAuth
}

func (m *Manager) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(m.auth)
}

func (m *Manager) Authenticator() func(http.Handler) http.Handler {
	return jwtauth.Authenticator(m.auth)
}

func (m *Manager) CreateToken(user_id string) (string, error) {
	claims := map[string]interface{}{
		"user_id": user_id,
		"exp":     time.Now().Add(time.Minute * 15),
	}
	_, token, err := m.auth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}
	return token, nil
}
