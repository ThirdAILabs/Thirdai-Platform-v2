package llm_dispatch

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	logger := slog.Default()
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid request", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "invalid request: " + err.Error()})
		return
	}

	// Validate required fields
	if req.Query == "" {
		logger.Error("missing required field: query")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Field required: query"})
		return
	}

	logger.Info("processing generation request",
		"provider", req.Provider,
		"model", req.Model,
	)

	if req.Key == "" {
		logger.Error("missing API key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "No generative AI key provided"})
		return
	}

	provider, err := NewLLMProvider(req.Provider, req.Key, logger)
	if err != nil {
		logger.Error("failed to create provider", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
		return
	}

	textChan, errChan := provider.Stream(&req)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"detail": "streaming unsupported"})
		return
	}

	var generatedResponse strings.Builder
	notify := r.Context().Done()

	for {
		select {
		case text, ok := <-textChan:
			if !ok {
				return
			}
			generatedResponse.WriteString(text)
			fmt.Fprintf(w, "data: %s\n\n", text)
			flusher.Flush()
		case err, ok := <-errChan:
			if !ok {
				return
			}
			logger.Error("error streaming response", "error", err)
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
			flusher.Flush()
			return
		case <-notify:
			logger.Info("client disconnected")
			return
		}
	}
}

func NewRouter() chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// CORS middleware to allow browser-based clients (e.g. test UI, admin dashboards) 
	// to make requests to this API. This is needed because browsers enforce the Same-Origin Policy
	// and require explicit CORS headers to allow cross-origin requests.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Routes
	r.Post("/llm-dispatch/generate", handleGenerate)
	r.Get("/llm-dispatch/health", handleHealth)

	return r
}