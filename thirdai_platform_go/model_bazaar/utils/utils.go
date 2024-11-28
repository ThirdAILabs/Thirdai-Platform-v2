package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
