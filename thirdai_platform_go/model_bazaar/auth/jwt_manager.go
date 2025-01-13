package auth

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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
	return jwtauth.Verifier(m.auth)
}

func (m *JwtManager) Authenticator() func(http.Handler) http.Handler {
	return jwtauth.Authenticator(m.auth)
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

func ModelIdsFromContext(r *http.Request) ([]uuid.UUID, error) {
	modelIDsStr, err := ValueFromContext(r, "model_ids")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve model_ids from context: %w", err)
	}

	if strings.TrimSpace(modelIDsStr) == "" {
		return nil, fmt.Errorf("model_ids claim is empty")
	}

	modelIDStrs := strings.Split(modelIDsStr, ",")

	var modelIDs []uuid.UUID
	for _, idStr := range modelIDStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid model_id '%s': %w", idStr, err)
		}

		modelIDs = append(modelIDs, id)
	}

	if len(modelIDs) == 0 {
		return nil, fmt.Errorf("no valid model_ids found in claim")
	}

	return modelIDs, nil
}

func UserFromContext(r *http.Request) (schema.User, error) {
	userUntyped := r.Context().Value("user")
	if userUntyped == nil {
		return schema.User{}, fmt.Errorf("user field not found in request context")
	}
	user, ok := userUntyped.(schema.User)
	if !ok {
		return schema.User{}, fmt.Errorf("invalid value for user field")
	}
	return user, nil
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
