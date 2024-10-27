package storage

import (
	"path/filepath"
)

func ModelPath(modelId string) string {
	return filepath.Join("models", modelId)
}

func DataPath(modelId string) string {
	return filepath.Join("models", modelId)
}
