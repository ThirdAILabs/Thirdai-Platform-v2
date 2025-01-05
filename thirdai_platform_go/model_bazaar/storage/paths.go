package storage

import (
	"path/filepath"

	"github.com/google/uuid"
)

func ModelPath(modelId uuid.UUID) string {
	return filepath.Join("models", modelId.String())
}

func DataPath(modelId uuid.UUID) string {
	return filepath.Join("data", modelId.String())
}
