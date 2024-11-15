package config

import (
	"fmt"
	"thirdai_platform/model_bazaar/schema"
)

const (
	FileLocLocal = "local"
	FileLocS3    = "s3"
	FileLocAzure = "azure"
	FileLocGcp   = "gcp"
)

const (
	ndbDataType = "ndb"
	nlpDataType = "nlp"
)

type FileInfo struct {
	Path     string                 `json:"path"`
	Location string                 `json:"location"`
	DocId    *string                `json:"doc_id"`
	Options  map[string]interface{} `json:"options"`
	Metadata map[string]interface{} `json:"metadata"`
}

func validateFileInfo(files []FileInfo) error {
	for i, file := range files {
		if file.Path == "" {
			return fmt.Errorf("file path cannot be empty")
		}
		if file.Location != FileLocLocal && file.Location != FileLocS3 && file.Location != FileLocAzure && file.Location != FileLocGcp {
			return fmt.Errorf("invalid file location '%v', must be 'local', 's3', 'azure', or 'gcp'", file.Location)
		}
		if file.Options == nil {
			files[i].Options = map[string]interface{}{}
		}
	}
	return nil
}

type NdbOptions struct {
	ModelType      string `json:"model_type"`
	InMemory       bool   `json:"in_memory"`
	AdvancedSearch bool   `json:"advanced_search"`
}

func (opts *NdbOptions) Validate() error {
	opts.ModelType = schema.NdbModel
	return nil
}

type NDBData struct {
	ModelDataType string `json:"model_data_type"`

	UnsupervisedFiles []FileInfo `json:"unsupervised_files"`
	SupervisedFiles   []FileInfo `json:"supervised_files"`

	Deletions []string `json:"deletions"`
}

func (data *NDBData) Validate() error {
	data.ModelDataType = ndbDataType

	if data.UnsupervisedFiles == nil {
		data.UnsupervisedFiles = []FileInfo{}
	}
	if data.SupervisedFiles == nil {
		data.SupervisedFiles = []FileInfo{}
	}
	if data.Deletions == nil {
		data.Deletions = []string{}
	}

	if len(data.UnsupervisedFiles)+len(data.SupervisedFiles) == 0 {
		return fmt.Errorf("NDB training requires either supervised or unsupervised data")
	}

	if err := validateFileInfo(data.UnsupervisedFiles); err != nil {
		return fmt.Errorf("invalid unsupervised files: %w", err)
	}

	if err := validateFileInfo(data.SupervisedFiles); err != nil {
		return fmt.Errorf("invalid supervised files: %w", err)
	}

	return nil
}

type NlpTokenOptions struct {
	ModelType string `json:"model_type"`

	TargetLabels []string `json:"target_labels"`
	SourceColumn string   `json:"source_column"`
	TargetColumn string   `json:"target_column"`
	DefaultTag   string   `json:"default_tag"`
}

func (opts *NlpTokenOptions) Validate() error {
	opts.ModelType = schema.NlpTokenModel
	opts.DefaultTag = "O"

	if opts.SourceColumn == "" {
		return fmt.Errorf("source_column must be specified")
	}

	if opts.TargetColumn == "" {
		return fmt.Errorf("target_column must be specified")
	}

	return nil
}

type NlpTextOptions struct {
	ModelType string `json:"model_type"`

	TextColumn     string `json:"text_column"`
	LabelColumn    string `json:"label_column"`
	NTargetClasses int    `json:"n_target_classes"`
	Delimiter      string `json:"delimiter"`
}

func (opts *NlpTextOptions) Validate() error {
	opts.ModelType = schema.NlpTextModel
	opts.Delimiter = ","

	if opts.TextColumn == "" {
		return fmt.Errorf("text_column must be specified")
	}

	if opts.LabelColumn == "" {
		return fmt.Errorf("label_column must be specified")
	}

	if opts.NTargetClasses <= 0 {
		return fmt.Errorf("n_target_classes must be > 0")
	}

	return nil
}

type NlpData struct {
	ModelDataType string `json:"model_data_type"`

	SupervisedFiles []FileInfo `json:"supervised_files"`
	TestFiles       []FileInfo `json:"test_files"`
}

func (data *NlpData) Validate() error {
	data.ModelDataType = nlpDataType

	if data.SupervisedFiles == nil {
		data.SupervisedFiles = []FileInfo{}
	}
	if data.TestFiles == nil {
		data.TestFiles = []FileInfo{}
	}

	if len(data.SupervisedFiles) == 0 {
		return fmt.Errorf("Nlp training requires training files")
	}

	if err := validateFileInfo(data.SupervisedFiles); err != nil {
		return fmt.Errorf("invalid supervised files: %w", err)
	}

	if err := validateFileInfo(data.TestFiles); err != nil {
		return fmt.Errorf("invalid test files: %w", err)
	}

	return nil
}

type NlpTrainOptions struct {
	Epochs             int      `json:"epochs"`
	LearningRate       float32  `json:"learning_rate"`
	BatchSize          int      `json:"batch_size"`
	MaxInMemoryBatches *int     `json:"max_in_memory_batches"`
	TestSplit          *float32 `json:"test_split"`
}

func (opts *NlpTrainOptions) Validate() error {
	if opts.Epochs == 0 {
		opts.Epochs = 1
	}

	if opts.LearningRate == 0 {
		opts.LearningRate = 1e-4
	}

	if opts.BatchSize == 0 {
		opts.BatchSize = 2048
	}

	return nil
}

type TrainConfig struct {
	ModelId             string  `json:"model_id"`
	ModelType           string  `json:"model_type"`
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

func (opts *JobOptions) Validate() error {
	opts.AllocationCores = max(opts.AllocationCores, 1)
	if opts.AllocationMemory < 500 {
		opts.AllocationMemory = 6800
	}
	return nil
}

func (opts *JobOptions) CpuUsageMhz() int {
	return opts.AllocationCores * 2400
}
