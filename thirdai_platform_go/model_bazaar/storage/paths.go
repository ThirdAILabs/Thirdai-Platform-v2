package storage

import (
	"path/filepath"

	"github.com/google/uuid"
)

func ModelPath(modelId uuid.UUID) string {
	return filepath.Join("models", modelId.String())
}

func ModelMetadataPath(modelId uuid.UUID) string {
	return filepath.Join(ModelPath(modelId), "model", "metadata.json")
}

func DataPath(modelId uuid.UUID) string {
	return filepath.Join("data", modelId.String())
}

func UploadPath(id uuid.UUID) string {
	return filepath.Join("uploads", id.String())
}
