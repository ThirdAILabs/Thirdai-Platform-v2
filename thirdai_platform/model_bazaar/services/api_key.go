package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"thirdai_platform/model_bazaar/auth"
	"thirdai_platform/model_bazaar/schema"
	"thirdai_platform/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func removeApiKeyPrefix(input string) (string, error) {
	expectedPrefix := thirdaiPlatformKeyPrefix + "-"
	if strings.HasPrefix(input, expectedPrefix) {
		trimmed := strings.TrimPrefix(input, expectedPrefix)
		return trimmed, nil
	}
	return "", fmt.Errorf("input string must start with the prefix '%s-'", thirdaiPlatformKeyPrefix)
}

func generateApiKey() (string, string, error) {
	var secret string
	var secretHash string

	secret, err := generateRandomString(32)
	if err != nil {
		return "", "", err
	}

	secretHash = hashSecret(secret)

	fullKey := fmt.Sprintf("%s-%s", thirdaiPlatformKeyPrefix, secret)
	return fullKey, secretHash, nil
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	str := base64.RawURLEncoding.EncodeToString(bytes)
	if len(str) < n {
		return "", errors.New("insufficient length in generated string")
	}
	return str[:n], nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func validateApiKey(db *gorm.DB, r *http.Request) (uuid.UUID, time.Time, error) {
	fullKey := r.Header.Get("X-API-Key")

	if fullKey == "" {
		return uuid.Nil, time.Time{}, ErrMissingAPIKey
	}

	secret, err := removeApiKeyPrefix(fullKey)
	if err != nil {
		return uuid.Nil, time.Time{}, ErrInvalidAPIKey
	}

	hashedKey := hashSecret(secret)

	var record schema.UserAPIKey
	if err := db.Where("hashkey = ?", hashedKey).Preload("Models").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, time.Time{}, ErrInvalidAPIKey
		}
		return uuid.Nil, time.Time{}, fmt.Errorf("database error: %w", err)
	}

	if time.Now().After(record.ExpiryTime) {
		return uuid.Nil, time.Time{}, ErrExpiredAPIKey
	}

	if hashSecret(secret) != record.HashKey {
		return uuid.Nil, time.Time{}, ErrInvalidAPIKey
	}

	modelId, err := utils.URLParamUUID(r, "model_id")
	if err != nil {
		return uuid.Nil, time.Time{}, fmt.Errorf("invalid model_id parameter: %w", err)
	}

	if !record.AllModels {
		for _, model := range record.Models {
			if model.Id == modelId {
				return record.CreatedBy, record.ExpiryTime, nil
			}
		}
	} else {
		return record.CreatedBy, record.ExpiryTime, nil
	}

	return uuid.Nil, time.Time{}, ErrAPIKeyModelMismatch
}

func eitherUserOrApiKeyAuthMiddleware(
	db *gorm.DB,
	userAuthMiddlewares chi.Middlewares,
) func(http.Handler) http.Handler {

	userAuthChain := chi.Chain(userAuthMiddlewares...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")

			if apiKey != "" {
				userID, expiry, err := validateApiKey(db, r)

				if err != nil {
					switch {
					case errors.Is(err, ErrMissingAPIKey):
						http.Error(w, err.Error(), http.StatusBadRequest)
					case errors.Is(err, ErrInvalidAPIKey):
						http.Error(w, err.Error(), http.StatusUnauthorized)
					case errors.Is(err, ErrExpiredAPIKey), errors.Is(err, ErrAPIKeyModelMismatch):
						http.Error(w, err.Error(), http.StatusForbidden)
					default:
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
					return
				}

				if userID == uuid.Nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				user, err := schema.GetUser(userID, db)
				if err != nil {
					http.Error(w, fmt.Sprintf("unable to get user: %v", err), http.StatusInternalServerError)
					return
				}

				reqCtx := r.Context()
				reqCtx = context.WithValue(reqCtx, auth.UserRequestContextKey, user)
				reqCtx = context.WithValue(reqCtx, auth.ContextAPIKeyExpiry, expiry)

				next.ServeHTTP(w, r.WithContext(reqCtx))
				return
			}

			finalHandler := userAuthChain.Handler(next)

			finalHandler.ServeHTTP(w, r)
		})
	}
}
