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
	fmt.Printf("Extracting URL parameter with key: %s\n", key) // Print key being searched
	param := chi.URLParam(r, key)
	fmt.Printf("URL parameter retrieved: '%s'\n", param) // Print the retrieved parameter

	if len(param) == 0 {
		fmt.Printf("Error: missing {%v} URL parameter\n", key) // Print missing parameter error
		return uuid.UUID{}, fmt.Errorf("missing {%v} URL parameter", key)
	}

	fmt.Printf("Parsing UUID: '%s'\n", param) // Print before parsing UUID
	id, err := uuid.Parse(param)
	if err != nil {
		fmt.Printf("Error parsing UUID: '%s', error: %v\n", param, err) // Print parse error details
		return uuid.UUID{}, fmt.Errorf("invalid uuid '%v' provided: %w", param, err)
	}

	fmt.Printf("UUID successfully parsed: '%s'\n", id.String()) // Print success message
	return id, nil
}
