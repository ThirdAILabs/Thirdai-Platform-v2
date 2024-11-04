package auth

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type JwtManager struct {
	auth *jwtauth.JWTAuth
}

func NewJwtManager() *JwtManager {
	return &JwtManager{auth: jwtauth.New("HS256", getSecret(), nil)}
}

func (m *JwtManager) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(m.auth)
}

func (m *JwtManager) Authenticator() func(http.Handler) http.Handler {
	return jwtauth.Authenticator(m.auth)
}

const userIdKey = "user_id"

func (m *JwtManager) CreateToken(key, value string, exp time.Duration) (string, error) {
	claims := map[string]interface{}{
		key:   value,
		"exp": time.Now().Add(exp),
	}
	_, token, err := m.auth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}
	return token, nil
}

func (m *JwtManager) CreateUserJwt(userId string) (string, error) {
	return m.CreateToken(userIdKey, userId, 15*time.Minute)
}

func ValueFromContext(r *http.Request, key string) (string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", fmt.Errorf("error retrieving auth claims: %v", err)
	}

	valueUncasted, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("invalid token: unable to locate %v", key)
	}

	value, ok := valueUncasted.(string)
	if !ok {
		return "", fmt.Errorf("invalid token: %v is invalid type", key)
	}

	return value, nil
}

func UserIdFromContext(r *http.Request) (string, error) {
	return ValueFromContext(r, userIdKey)
}

func getSecret() []byte {
	// This is only used for jwt secrets, if the server restarts the only issue is any
	// tokens issued before the restart (that aren't yet expired) will be invalidated.
	b := make([]byte, 16)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b
}
