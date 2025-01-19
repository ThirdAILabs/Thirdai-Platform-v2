package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParseRequestBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dest)
	if err != nil {
		slog.Error("error parsing request body", "error", err)
		http.Error(w, fmt.Sprintf("error parsing request body: %v", err), http.StatusBadRequest)
		return false
	}
	return true
}

func WriteJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("error serializing response body", "error", err)
		http.Error(w, fmt.Sprintf("error serializing response body: %v", err), http.StatusInternalServerError)
	}
}

func WriteSuccess(w http.ResponseWriter) {
	WriteJsonResponse(w, struct{}{})
}

func URLParam(r *http.Request, key string) (string, error) {
	param := chi.URLParam(r, key)
	if len(param) == 0 {
		return "", fmt.Errorf("missing {%v} url parameter", key)
	}
	return param, nil
}

func URLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)
	if len(param) == 0 {
		return uuid.Nil, fmt.Errorf("missing {%v} url parameter", key)
	}

	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid '%v' provided: %w", param, err)
	}

	return id, nil
}
