package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/url"
	"thirdai_platform/model_bazaar/config"
	"thirdai_platform/model_bazaar/services"
)

type PlatformClient struct {
	baseUrl   string
	authToken string
	userId    string
}

func New(baseUrl string) *PlatformClient {
	return &PlatformClient{baseUrl: baseUrl}
}

func (c *PlatformClient) Signup(username, email, password string) error {
	body, err := json.Marshal(map[string]string{
		"email": email, "username": username, "password": password,
	})
	if err != nil {
		return fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/user/signup")
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[map[string]string](u, body, nil, c.authToken)

	return err
}

func (c *PlatformClient) Login(email, password string) error {
	body, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/user/login")
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	data, err := post[map[string]string](u, body, nil, c.authToken)
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

	u, err := url.JoinPath(c.baseUrl, "/api/v2/train/upload-data")
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	headers := authHeader(c.authToken)
	headers["Content-Type"] = writer.FormDataContentType()

	res, err := postWithHeaders[map[string]string](u, body.Bytes(), nil, headers)
	if err != nil {
		return nil, err
	}

	artifactDir := res["artifact_path"]

	return updateLocalFilePrefixes(files, artifactDir), nil
}

func (c *PlatformClient) TrainNdb(name string, unsupervised []config.FileInfo, supervised []config.FileInfo) (*NdbClient, error) {
	return c.TrainNdbWithJobOptions(name, unsupervised, supervised, config.JobOptions{})
}

func (c *PlatformClient) TrainNdbWithJobOptions(name string, unsupervised []config.FileInfo, supervised []config.FileInfo, jobOptions config.JobOptions) (*NdbClient, error) {
	unsupervisedFiles, err := c.uploadFiles(unsupervised)
	if err != nil {
		return nil, fmt.Errorf("error uploading unsupervised files for training: %w", err)
	}

	supervisedFiles, err := c.uploadFiles(supervised)
	if err != nil {
		return nil, fmt.Errorf("error uploading supervised files for training: %w", err)
	}

	params := services.NdbTrainOptions{
		ModelName:    name,
		ModelOptions: &config.NdbOptions{},
		Data: config.NDBData{
			UnsupervisedFiles: unsupervisedFiles,
			SupervisedFiles:   supervisedFiles,
		},
		JobOptions: jobOptions,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/train/ndb")
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[map[string]string](u, body, nil, c.authToken)
	if err != nil {
		return nil, err
	}

	return &NdbClient{
		ModelClient{
			baseUrl:   c.baseUrl,
			authToken: c.authToken,
			modelId:   res["model_id"],
		},
	}, nil
}

func (c *PlatformClient) TrainNlpToken(name string, labels []string, files []config.FileInfo) (*NlpTokenClient, error) {
	uploadFiles, err := c.uploadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("error uploading files for training: %w", err)
	}

	params := services.NlpTokenTrainOptions{
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
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/train/nlp-token")
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[map[string]string](u, body, nil, c.authToken)
	if err != nil {
		return nil, err
	}

	return &NlpTokenClient{
		ModelClient{
			baseUrl:   c.baseUrl,
			authToken: c.authToken,
			modelId:   res["model_id"],
		},
	}, nil
}

func (c *PlatformClient) TrainNlpText(name string, nTargetClasses int, files []config.FileInfo) (*NlpTextClient, error) {
	uploadFiles, err := c.uploadFiles(files)
	if err != nil {
		return nil, fmt.Errorf("error uploading files for training: %w", err)
	}

	params := services.NlpTextTrainOptions{
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
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/train/nlp-text")
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[map[string]string](u, body, nil, c.authToken)
	if err != nil {
		return nil, err
	}

	return &NlpTextClient{
		ModelClient{
			baseUrl:   c.baseUrl,
			authToken: c.authToken,
			modelId:   res["model_id"],
		},
	}, nil
}
