package client

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/services"
	"time"

	"github.com/google/uuid"
)

type PlatformClient struct {
	BaseClient
	userId string
}

func New(baseUrl string) *PlatformClient {
	return &PlatformClient{BaseClient: BaseClient{baseUrl: baseUrl}}
}

func (c *PlatformClient) Signup(username, email, password string) error {
	body := map[string]string{
		"email": email, "username": username, "password": password,
	}

	return c.Post("/api/v2/user/signup").Json(body).Do(nil)
}

func (c *PlatformClient) Login(email, password string) error {
	var data map[string]string
	err := c.Get("/api/v2/user/login").Login(email, password).Do(&data)
	if err != nil {
		return err
	}

	c.authToken = data["access_token"]
	c.userId = data["user_id"]

	return nil
}

func (c *PlatformClient) UseApiKey(api_key string) error {

	c.apiKey = api_key
	c.authToken = ""
	c.userId = ""
	return nil
}

func (c *PlatformClient) CreateAPIKey(modelIDs []uuid.UUID, name string, expiry time.Time, allModels bool) (string, error) {
	requestBody := map[string]interface{}{
		"model_ids":  modelIDs,
		"name":       name,
		"exp":        expiry,
		"all_models": allModels,
	}

	var response struct {
		ApiKey string `json:"api_key"`
	}

	err := c.Post("/api/v2/model/create-api-key").Json(requestBody).Do(&response)
	if err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}

	return response.ApiKey, nil
}

func (c *PlatformClient) ListAPIKeys() ([]services.APIKeyResponse, error) {
	var response []services.APIKeyResponse

	err := c.Get("/api/v2/model/list-api-keys").Do(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	return response, nil
}

func (c *PlatformClient) DeleteAPIKey(apiKeyID uuid.UUID) error {
	body := map[string]uuid.UUID{
		"api_key_id": apiKeyID,
	}

	err := c.Post("/api/v2/model/delete-api-key").Json(body).Do(nil)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

type FileInfo struct {
	Path     string                 `json:"path"`
	Location string                 `json:"location"`
	SourceId *string                `json:"source_id"`
	Options  map[string]interface{} `json:"options"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (c *PlatformClient) uploadFiles(files []FileInfo, subDir string) ([]config.TrainFile, error) {
	updatedFiles := make([]config.TrainFile, 0)

	for _, file := range files {
		if file.Location == "upload" || file.Location == "local" {
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			err := addFilesToMultipart(writer, []FileInfo{file})
			if err != nil {
				return nil, err
			}

			err = writer.Close()
			if err != nil {
				return nil, fmt.Errorf("error closing mutlipart writer: %w", err)
			}

			var res map[string]string
			err = c.Post("/api/v2/train/upload-data").Param("sub_dir", subDir).Header("Content-Type", writer.FormDataContentType()).Body(body).Do(&res)
			if err != nil {
				return nil, err
			}

			updatedFiles = append(updatedFiles, config.TrainFile{
				Location: "upload",
				Path:     res["upload_id"],
				SourceId: file.SourceId,
				Options:  file.Options,
				Metadata: file.Metadata,
			})
		} else {
			updatedFiles = append(updatedFiles, config.TrainFile{
				Path:     file.Path,
				Location: file.Location,
				SourceId: file.SourceId,
				Options:  file.Options,
				Metadata: file.Metadata,
			})
		}
	}

	return updatedFiles, nil
}

func (c *PlatformClient) TrainNdb(name string, unsupervised []FileInfo, supervised []FileInfo, jobOptions config.JobOptions) (*NdbClient, error) {
	return c.TrainNdbWithBaseModel(name, nil, unsupervised, supervised, jobOptions, false, nil)
}

func (c *PlatformClient) TrainNdbWithGenerativeSupervision(name string, unsupervised []FileInfo, supervised []FileInfo, jobOptions config.JobOptions, llmConfig *config.LLMConfig) (*NdbClient, error) {
	return c.TrainNdbWithBaseModel(name, nil, unsupervised, supervised, jobOptions, true, llmConfig)
}

func (c *PlatformClient) TrainNdbWithBaseModel(name string, baseModel *NdbClient, unsupervised []FileInfo, supervised []FileInfo, jobOptions config.JobOptions, generativeSupervision bool, llmConfig *config.LLMConfig) (*NdbClient, error) {
	unsupervisedFiles, err := c.uploadFiles(unsupervised, "")
	if err != nil {
		return nil, fmt.Errorf("error uploading unsupervised files for training: %w", err)
	}

	supervisedFiles, err := c.uploadFiles(supervised, "")
	if err != nil {
		return nil, fmt.Errorf("error uploading supervised files for training: %w", err)
	}

	var baseModelId *uuid.UUID
	if baseModel != nil {
		baseModelId = &baseModel.modelId
	}
	var modelOptions *config.NdbOptions
	if baseModel == nil {
		modelOptions = &config.NdbOptions{}
	}

	body := services.NdbTrainRequest{
		ModelName:    name,
		BaseModelId:  baseModelId,
		ModelOptions: modelOptions,
		Data: config.NDBData{
			UnsupervisedFiles: unsupervisedFiles,
			SupervisedFiles:   supervisedFiles,
		},
		JobOptions:            jobOptions,
		LLMConfig:             llmConfig,
		GenerativeSupervision: generativeSupervision,
	}

	var res newModelResponse
	err = c.Post("/api/v2/train/ndb").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NdbClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) TrainNlpToken(name string, labels []string, files []FileInfo, trainOptions config.NlpTrainOptions) (*NlpTokenClient, error) {
	return c.trainNlpTokenHelper(name, nil, labels, files, trainOptions)
}

func (c *PlatformClient) TrainNlpTokenWithBaseModel(name string, baseModel *NlpTokenClient, files []FileInfo, trainOptions config.NlpTrainOptions) (*NlpTokenClient, error) {
	return c.trainNlpTokenHelper(name, baseModel, nil, files, trainOptions)
}

func (c *PlatformClient) trainNlpTokenHelper(name string, baseModel *NlpTokenClient, labels []string, files []FileInfo, trainOptions config.NlpTrainOptions) (*NlpTokenClient, error) {
	uploadFiles, err := c.uploadFiles(files, "")
	if err != nil {
		return nil, fmt.Errorf("error uploading files for training: %w", err)
	}

	var baseModelId *uuid.UUID
	if baseModel != nil {
		baseModelId = &baseModel.modelId
	}
	var modelOptions *config.NlpTokenOptions
	if baseModel == nil {
		modelOptions = &config.NlpTokenOptions{
			TargetLabels: labels,
			SourceColumn: "source",
			TargetColumn: "target",
			DefaultTag:   "O",
		}
	}

	body := services.NlpTokenTrainRequest{
		ModelName:    name,
		BaseModelId:  baseModelId,
		ModelOptions: modelOptions,
		Data: config.NlpData{
			SupervisedFiles: uploadFiles,
		},
		TrainOptions: trainOptions,
	}

	var res newModelResponse
	err = c.Post("/api/v2/train/nlp-token").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTokenClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) TrainNlpText(name string, nTargetClasses int, files []FileInfo, trainOptions config.NlpTrainOptions) (*NlpTextClient, error) {
	return c.trainNlpTextHelper(name, nil, nTargetClasses, files, trainOptions)
}

func (c *PlatformClient) TrainNlpTextWithBaseModel(name string, baseModel *NlpTextClient, files []FileInfo, trainOptions config.NlpTrainOptions) (*NlpTextClient, error) {
	return c.trainNlpTextHelper(name, baseModel, -1, files, trainOptions)
}

func (c *PlatformClient) trainNlpTextHelper(name string, baseModel *NlpTextClient, nTargetClasses int, files []FileInfo, trainOptions config.NlpTrainOptions) (*NlpTextClient, error) {
	uploadFiles, err := c.uploadFiles(files, "")
	if err != nil {
		return nil, fmt.Errorf("error uploading files for training: %w", err)
	}

	var baseModelId *uuid.UUID
	if baseModel != nil {
		baseModelId = &baseModel.modelId
	}
	var modelOptions *config.NlpTextOptions
	if baseModel == nil {
		modelOptions = &config.NlpTextOptions{
			NTargetClasses: nTargetClasses,
			TextColumn:     "text",
			LabelColumn:    "labels",
			Delimiter:      ",",
		}
	}

	body := services.NlpTextTrainRequest{
		ModelName:    name,
		BaseModelId:  baseModelId,
		ModelOptions: modelOptions,
		Data: config.NlpData{
			SupervisedFiles: uploadFiles,
		},
		TrainOptions: trainOptions,
	}

	var res newModelResponse
	err = c.Post("/api/v2/train/nlp-text").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTextClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) TrainNlpDoc(name string, directory string, trainOptions config.NlpTrainOptions) (*NlpTextClient, error) {
	categories := map[string][]FileInfo{}

	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		category := filepath.Base(filepath.Dir(path))

		if _, ok := categories[category]; !ok {
			categories[category] = make([]FileInfo, 0)
		}
		categories[category] = append(categories[category], FileInfo{Path: path, Location: "upload"})

		return nil
	})

	if err != nil {
		return nil, err
	}

	allFiles := []config.TrainFile{}
	for category, files := range categories {
		uploadFiles, err := c.uploadFiles(files, category)
		if err != nil {
			return nil, fmt.Errorf("error uploading files for training: %w", err)
		}

		allFiles = append(allFiles, uploadFiles...)
	}

	body := services.NlpTextTrainRequest{
		ModelName:         name,
		BaseModelId:       nil,
		DocClassification: true,
		ModelOptions: &config.NlpTextOptions{
			NTargetClasses: len(categories),
		},
		Data: config.NlpData{
			SupervisedFiles: allFiles,
		},
		TrainOptions: trainOptions,
	}

	var res newModelResponse
	err = c.Post("/api/v2/train/nlp-text").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTextClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) TrainNlpTokenDatagen(name string, taskPrompt string, options config.NlpTokenDatagenOptions, trainOptions config.NlpTrainOptions) (*NlpTokenClient, error) {
	body := services.NlpTrainDatagenRequest{
		ModelName:    name,
		TaskPrompt:   taskPrompt,
		TokenOptions: &options,
		TrainOptions: trainOptions,
	}

	var res newModelResponse
	err := c.Post("/api/v2/train/nlp-datagen").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTokenClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) TrainNlpTextDatagen(name string, taskPrompt string, options config.NlpTextDatagenOptions, trainOptions config.NlpTrainOptions) (*NlpTextClient, error) {
	body := services.NlpTrainDatagenRequest{
		ModelName:    name,
		TaskPrompt:   taskPrompt,
		TextOptions:  &options,
		TrainOptions: trainOptions,
	}

	var res newModelResponse
	err := c.Post("/api/v2/train/nlp-datagen").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTextClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) CreateEnterpriseSearchWorkflow(modelName string, retrieval *NdbClient, guardrail *NlpTokenClient) (*EnterpriseSearchClient, error) {
	var guardrailId *uuid.UUID = nil
	if guardrail != nil {
		guardrailId = &guardrail.modelId
	}

	body := services.EnterpriseSearchRequest{
		ModelName:   modelName,
		RetrievalId: retrieval.modelId,
		GuardrailId: guardrailId,
	}

	var res newModelResponse
	err := c.Post("/api/v2/workflow/enterprise-search").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &EnterpriseSearchClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

func (c *PlatformClient) CreateKnowledgeExtractionWorkflow(modelName string, questions []string, llmProvider string, generateAnswers bool) (*KnowledgeExtractionClient, error) {
	questionKeywords := make([]services.QuestionKeywords, 0, len(questions))
	for _, q := range questions {
		questionKeywords = append(questionKeywords, services.QuestionKeywords{Question: q})
	}

	body := services.KnowledgeExtractionRequest{
		ModelName:       modelName,
		Questions:       questionKeywords,
		LlmProvider:     llmProvider,
		GenerateAnswers: &generateAnswers,
	}

	var res newModelResponse
	err := c.Post("/api/v2/workflow/knowledge-extraction").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &KnowledgeExtractionClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.ModelId,
		},
	}, nil
}

type uploadResponse struct {
	ModelId   uuid.UUID `json:"model_id"`
	ModelType string    `json:"model_type"`
}

func (c *PlatformClient) UploadModel(modelName, path string) (interface{}, error) {
	modelFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error reading model file %v: %w", path, err)
	}
	defer modelFile.Close()

	body := map[string]string{"model_name": modelName}
	var uploadToken map[string]string
	err = c.Post("/api/v2/model/upload").Json(body).Do(&uploadToken)
	if err != nil {
		return nil, fmt.Errorf("error starting model upload: %w", err)
	}

	chunkIdx := 0
	chunk := make([]byte, 10*1024*1024)
	for {
		n, rerr := modelFile.Read(chunk)
		if rerr != nil && rerr != io.EOF {
			return nil, fmt.Errorf("error reading from model file: %w", err)
		}

		err := c.Post(fmt.Sprintf("/api/v2/model/upload/%d", chunkIdx)).
			Auth(uploadToken["token"]).
			Body(bytes.NewReader(chunk[:n])).Do(nil)
		if err != nil {
			return nil, fmt.Errorf("error sending chunk %d: %w", chunkIdx, err)
		}
		if rerr == io.EOF {
			break
		}
		chunkIdx++
	}

	var res uploadResponse
	err = c.Post("/api/v2/model/upload/commit").Auth(uploadToken["token"]).Do(&res)
	if err != nil {
		return nil, fmt.Errorf("error committing model upload: %w", err)
	}

	modelClient := ModelClient{BaseClient: c.BaseClient, modelId: res.ModelId}

	switch res.ModelType {
	case "ndb":
		return &NdbClient{modelClient}, nil
	case "nlp-text", "nlp-doc":
		return &NlpTextClient{modelClient}, nil
	case "nlp-token":
		return &NlpTokenClient{modelClient}, nil
	case "ke":
		return &KnowledgeExtractionClient{modelClient}, nil
	default:
		return nil, fmt.Errorf("invalid model type: %v", res.ModelType)
	}
}

func (c *PlatformClient) Backup(config services.BackupRequest) error {
	return c.Post("/api/v2/recovery/backup").Json(config).Do(nil)
}

func (c *PlatformClient) LocalBackups() ([]string, error) {
	var backups []string
	err := c.Get("/api/v2/recovery/backups").Do(&backups)
	return backups, err
}
