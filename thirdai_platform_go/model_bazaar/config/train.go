package config

import (
	"fmt"
	"thirdai_platform/model_bazaar/schema"
)

const (
	ModelTypeNdb              = "ndb"
	ModelTypeUdt              = "udt"
	ModelTypeEnterpriseSearch = "enterprise-search"
)

const (
	ModelDataTypeNdb        = "ndb"
	ModelDataTypeUdt        = "udt"
	ModelDataTypeUdtDatagen = "udt_datagen"
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
	ModelType  string                 `json:"model_type"`
	NdbOptions map[string]interface{} `json:"ndb_options"`
}

func (o *NdbOptions) SubType() (string, error) {
	subtype, ok := o.NdbOptions["ndb_subtype"]
	if !ok {
		return schema.NdbV2Subtype, nil
	}
	subtypeStr, ok := subtype.(string)
	if !ok {
		return "", fmt.Errorf("field 'ndb_subtype' must be a string")
	}
	if subtypeStr != schema.NdbV1Subtype && subtypeStr != schema.NdbV2Subtype {
		return "", fmt.Errorf("invalid 'ndb_subtype' %v, must be 'v1' or 'v2'", subtypeStr)
	}
	return subtypeStr, nil
}

type NDBData struct {
	ModelDataType string `json:"model_data_type"`

	UnsupervisedFiles []FileInfo `json:"unsupervised_files"`
	SupervisedFiles   []FileInfo `json:"supervised_files"`
	TestFiles         []FileInfo `json:"test_files"`

	Deletions []string `json:"deletions"`
}

type TrainConfig struct {
	ModelBazaarDir      string  `json:"model_bazaar_dir"`
	LicenseKey          string  `json:"license_key"`
	ModelBazaarEndpoint string  `json:"model_bazaar_endpoint"`
	UpdateToken         string  `json:"update_token"`
	ModelId             string  `json:"model_id"`
	BaseModelId         *string `json:"base_model_id"`

	ModelOptions interface{} `json:"model_options"`
	Data         interface{} `json:"data"`

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
