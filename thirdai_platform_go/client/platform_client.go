package client

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/services"
)

type PlatformClient struct {
	baseClient
	userId string
}

func New(baseUrl string) *PlatformClient {
	return &PlatformClient{baseClient: baseClient{baseUrl: baseUrl}}
}

func (c *PlatformClient) Signup(username, email, password string) error {
	body := map[string]string{
		"email": email, "username": username, "password": password,
	}

	return c.Post("/api/v2/user/signup").Json(body).Do(nil)
}

func (c *PlatformClient) Login(email, password string) error {
	body := map[string]string{"email": email, "password": password}

	var data map[string]string
	err := c.Post("/api/v2/user/login").Json(body).Do(&data)
	if err != nil {
		return err
	}

	c.authToken = data["access_token"]
	c.userId = data["user_id"]

	return nil
}

func (c *PlatformClient) uploadFiles(files []config.FileInfo) ([]config.FileInfo, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	err := addFilesToMultipart(writer, files)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing mutlipart writer: %w", err)
	}

	var res map[string]string
	err = c.Post("/api/v2/train/upload-data").Header("Content-Type", writer.FormDataContentType()).Body(body).Do(&res)
	if err != nil {
		return nil, err
	}

	artifactDir := res["artifact_path"]

	return updateLocalFilePrefixes(files, artifactDir), nil
}

func (c *PlatformClient) TrainNdb(name string, unsupervised []config.FileInfo, supervised []config.FileInfo, jobOptions config.JobOptions) (*NdbClient, error) {
	unsupervisedFiles, err := c.uploadFiles(unsupervised)
	if err != nil {
		return nil, fmt.Errorf("error uploading unsupervised files for training: %w", err)
	}

	supervisedFiles, err := c.uploadFiles(supervised)
	if err != nil {
		return nil, fmt.Errorf("error uploading supervised files for training: %w", err)
	}

	body := services.NdbTrainRequest{
		ModelName:    name,
		ModelOptions: &config.NdbOptions{},
		Data: config.NDBData{
			UnsupervisedFiles: unsupervisedFiles,
			SupervisedFiles:   supervisedFiles,
		},
		JobOptions: jobOptions,
	}

	var res map[string]string
	err = c.Post("/api/v2/train/ndb").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NdbClient{
		ModelClient{
			baseClient: c.baseClient,
			modelId:    res["model_id"],
		},
	}, nil
}

func (c *PlatformClient) TrainNlpToken(name string, labels []string, files []config.FileInfo, trainOptions config.NlpTrainOptions) (*NlpTokenClient, error) {
	uploadFiles, err := c.uploadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("error uploading files for training: %w", err)
	}

	body := services.NlpTokenTrainRequest{
		ModelName: name,
		ModelOptions: &config.NlpTokenOptions{
			TargetLabels: labels,
			SourceColumn: "source",
			TargetColumn: "target",
			DefaultTag:   "O",
		},
		Data: config.NlpData{
			SupervisedFiles: uploadFiles,
		},
		TrainOptions: trainOptions,
	}

	var res map[string]string
	err = c.Post("/api/v2/train/nlp-token").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTokenClient{
		ModelClient{
			baseClient: c.baseClient,
			modelId:    res["model_id"],
		},
	}, nil
}

func (c *PlatformClient) TrainNlpText(name string, nTargetClasses int, files []config.FileInfo, trainOptions config.NlpTrainOptions) (*NlpTextClient, error) {
	uploadFiles, err := c.uploadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("error uploading files for training: %w", err)
	}

	body := services.NlpTextTrainRequest{
		ModelName: name,
		ModelOptions: &config.NlpTextOptions{
			NTargetClasses: nTargetClasses,
			TextColumn:     "text",
			LabelColumn:    "labels",
			Delimiter:      ",",
		},
		Data: config.NlpData{
			SupervisedFiles: uploadFiles,
		},
		TrainOptions: trainOptions,
	}

	var res map[string]string
	err = c.Post("/api/v2/train/nlp-text").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTextClient{
		ModelClient{
			baseClient: c.baseClient,
			modelId:    res["model_id"],
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

	var res map[string]string
	err := c.Post("/api/v2/train/nlp-datagen").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTokenClient{
		ModelClient{
			baseClient: c.baseClient,
			modelId:    res["model_id"],
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

	var res map[string]string
	err := c.Post("/api/v2/train/nlp-datagen").Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NlpTextClient{
		ModelClient{
			baseClient: c.baseClient,
			modelId:    res["model_id"],
		},
	}, nil
}
