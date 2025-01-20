package main

import (
	"log/slog"
	"os"
	"thirdai_platform/model_bazaar/jobs/llm_dispatch"
)

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create and start server
	server := llm_dispatch.NewServer()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	logger.Info("starting LLM dispatch service", "port", port)
	if err := server.Run(":" + port); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
} 