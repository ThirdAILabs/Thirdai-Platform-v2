package config

const (
	ModelTypeNdb              = "ndb"
	ModelTypeUdt              = "udt"
	ModelTypeEnterpriseSearch = "enterprise-search"
)

const (
	FileLocLocal = "local"
	FileLocS3    = "s3"
)

type FileInfo struct {
	Path     string                  `json:"path"`
	Location string                  `json:"location"`
	DocId    *string                 `json:"doc_id"`
	Options  map[string]interface{}  `json:"options"`
	Metadata *map[string]interface{} `json:"metadata"`
}

type NdbOptions struct {
	NdbOptions map[string]interface{} `json:"ndb_options"`
}

type NDBData struct {
	UnsupervisedFiles []FileInfo `json:"unsupervised_files"`
	SupervisedFiles   []FileInfo `json:"supervised_files"`
	TestFiles         []FileInfo `json:"test_files"`

	Deletions []string `json:"deletions"`
}

type TrainConfig struct {
	ModelId             string  `json:"model_id"`
	ModelBazaarDir      string  `json:"model_bazaar_dir"`
	ModelBazaarEndpoint string  `json:"model_bazaar_endpoint"`
	JobAuthToken        string  `json:"job_auth_token"`
	LicenseKey          string  `json:"license_key"`
	BaseModelId         *string `json:"base_model_id"`

	ModelOptions interface{} `json:"model_options"`
	Data         interface{} `json:"data"`
	TrainOptions interface{} `json:"train_options"`

	JobOptions JobOptions `json:"job_options"`

	IsRetraining bool `json:"is_retraining"`
}

type JobOptions struct {
	AllocationCores  int `json:"allocation_cores"`
	AllocationMemory int `json:"allocation_memory"`
}

func (j *JobOptions) EnsureValid() {
	j.AllocationCores = max(j.AllocationCores, 1)
	if j.AllocationMemory < 500 {
		j.AllocationMemory = 6800
	}
}

func (j *JobOptions) CpuUsageMhz() int {
	return j.AllocationCores * 2400
}
