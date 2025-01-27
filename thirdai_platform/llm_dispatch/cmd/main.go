package main

import (
	"log/slog"
	"net/http"
	"os"
	"thirdai_platform/llm_dispatch"
)

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := "8000"
	logger.Info("starting LLM generation service", "port", port)
	
	router := llm_dispatch.NewRouter()
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
} 