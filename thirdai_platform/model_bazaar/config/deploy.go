package config

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
)

type DeployConfig struct {
	ModelId             uuid.UUID         `json:"model_id"`
	UserId              uuid.UUID         `json:"user_id"`
	ModelType           string            `json:"model_type"`
	ModelBazaarDir      string            `json:"model_bazaar_dir"`
	HostDir             string            `json:"host_dir"`
	ModelBazaarEndpoint string            `json:"model_bazaar_endpoint"`
	LicenseKey          string            `json:"license_key"`
	JobAuthToken        string            `json:"job_auth_token"`
	Autoscaling         bool              `json:"autoscaling_enabled"`
	Options             map[string]string `json:"options"`
}

func LoadDeployConfig(configPath string) (*DeployConfig, error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config DeployConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	return &config, nil
}
