package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/schema"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type JwtManager struct {
	auth *jwtauth.JWTAuth
}

func NewJwtManager(secret []byte) *JwtManager {
	return &JwtManager{auth: jwtauth.New("HS256", secret, nil)}
}

func (m *JwtManager) Verifier() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verifier(m.auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}))
	}
}

func (m *JwtManager) Authenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Authenticator(m.auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}))
	}
}

const (
	userIdKey  = "user_id"
	modelIdKey = "model_id"
)

func (m *JwtManager) createToken(key, value string, exp time.Duration) (string, error) {
	claims := map[string]interface{}{
		key:   value,
		"exp": time.Now().Add(exp),
	}
	_, token, err := m.auth.Encode(claims)
	if err != nil {
		slog.Error("error generating jwt", "error", err)
		return "", fmt.Errorf("error generating access token: %w", err)
	}
	return token, nil
}

func (m *JwtManager) CreateUserJwt(userId uuid.UUID) (string, error) {
	return m.createToken(userIdKey, userId.String(), 15*time.Minute)
}

func (m *JwtManager) CreateModelJwt(modelId uuid.UUID, exp time.Duration) (string, error) {
	return m.createToken(modelIdKey, modelId.String(), exp)
}

func ValueFromContext(r *http.Request, key string) (string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", fmt.Errorf("error retrieving auth claims: %w", err)
	}

	valueUncasted, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("invalid token: unable to locate key %v in claims", key)
	}

	value, ok := valueUncasted.(string)
	if !ok {
		return "", fmt.Errorf("invalid token: value for key %v has invalid type", key)
	}

	return value, nil
}

func ModelIdFromContext(r *http.Request) (uuid.UUID, error) {
	value, err := ValueFromContext(r, modelIdKey)
	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid uuid '%v' provided: %w", value, err)
	}
	return id, nil
}

func UserFromContext(r *http.Request) (schema.User, error) {
	userUntyped := r.Context().Value(UserRequestContextKey)
	if userUntyped == nil {
		return schema.User{}, fmt.Errorf("user field not found in request context")
	}
	user, ok := userUntyped.(schema.User)
	if !ok {
		return schema.User{}, fmt.Errorf("invalid value for user field")
	}
	return user, nil
}

func GetAPIKeyExpiry(ctx context.Context) (time.Time, bool) {
	expiry, ok := ctx.Value(ContextAPIKeyExpiry).(time.Time)
	return expiry, ok
}
