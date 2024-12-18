package config

type DeployConfig struct {
	ModelId             string            `json:"model_id"`
	UserId              string            `json:"user_id"`
	ModelType           string            `json:"model_type"`
	ModelBazaarDir      string            `json:"model_bazaar_dir"`
	HostDir             string            `json:"host_dir"`
	ModelBazaarEndpoint string            `json:"model_bazaar_endpoint"`
	LicenseKey          string            `json:"license_key"`
	JobAuthToken        string            `json:"job_auth_token"`
	Autoscaling         bool              `json:"autoscaling_enabled"`
	Options             map[string]string `json:"options"`
}
