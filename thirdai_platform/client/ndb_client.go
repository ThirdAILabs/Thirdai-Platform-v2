package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"thirdai_platform/model_bazaar/services"

	"github.com/google/uuid"
)

type NdbClient struct {
	ModelClient
}

type ndbSearchParams struct {
	Query string `json:"query"`
	Topk  int    `json:"top_k"`
}

type NdbSearchResult struct {
	Id      int                    `json:"id"`
	Text    string                 `json:"text"`
	Context string                 `json:"context"`
	Source  string                 `json:"source"`
	Metdata map[string]interface{} `json:"metadata"`
	Score   float32                `json:"score"`
}

type ndbSearchResults struct {
	References []NdbSearchResult `json:"references"`
}

func (c *NdbClient) Search(query string, topk int) ([]NdbSearchResult, error) {
	body := ndbSearchParams{Query: query, Topk: topk}

	var res wrappedData[ndbSearchResults]
	err := c.Post(fmt.Sprintf("/%v/search", c.deploymentId())).Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return res.Data.References, nil
}

type insertParams struct {
	Documents []FileInfo `json:"documents"`
}

func (c *NdbClient) Insert(files []FileInfo) error {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	err := addFilesToMultipart(writer, files)
	if err != nil {
		return err
	}

	part, err := writer.CreateFormField("documents")
	if err != nil {
		return fmt.Errorf("error creating request part: %w", err)
	}

	params := insertParams{Documents: files}
	for i := range params.Documents {
		if params.Documents[i].Location == "upload" {
			// NDB deployment doesn't support upload yet
			params.Documents[i].Location = "local"
		}
		if params.Documents[i].Options == nil {
			params.Documents[i].Options = map[string]interface{}{}
		}
	}
	err = json.NewEncoder(part).Encode(params)
	if err != nil {
		return fmt.Errorf("error encoding request: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing mutlipart writer: %w", err)
	}

	return c.Post(fmt.Sprintf("/%v/insert", c.deploymentId())).Header("Content-Type", writer.FormDataContentType()).Body(body).Do(nil)
}

type deleteParams struct {
	SourceIds []string `json:"source_ids"`
}

func (c *NdbClient) DeleteDocs(doc_ids []string) error {
	body := deleteParams{SourceIds: doc_ids}

	return c.Post(fmt.Sprintf("/%v/delete", c.deploymentId())).Json(body).Do(nil)
}

type UpvotePair struct {
	QueryText     string `json:"query_text"`
	ReferenceId   int    `json:"reference_id"`
	ReferenceText string `json:"reference_text"`
}

type upvoteParams struct {
	TextIdPairs []UpvotePair `json:"text_id_pairs"`
}

func (c *NdbClient) Upvote(samples []UpvotePair) error {
	body := upvoteParams{TextIdPairs: samples}

	return c.Post(fmt.Sprintf("/%v/upvote", c.deploymentId())).Json(body).Do(nil)
}

type AssociatePair struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type associateParams struct {
	TextPairs []AssociatePair `json:"text_pairs"`
}

func (c *NdbClient) Associate(samples []AssociatePair) error {
	body := associateParams{TextPairs: samples}

	return c.Post(fmt.Sprintf("/%v/associate", c.deploymentId())).Json(body).Do(nil)
}

type Source struct {
	Source   string `json:"source"`
	SourceId string `json:"source_id"`
	Version  int    `json:"version"`
}

func (c *NdbClient) Sources() ([]Source, error) {
	var res wrappedData[[]Source]
	err := c.Get(fmt.Sprintf("/%v/sources", c.deploymentId())).Do(&res)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *NdbClient) Retrain(newModelName string) (*NdbClient, error) {
	body := services.NdbRetrainRequest{
		ModelName:   newModelName,
		BaseModelId: c.modelId,
	}

	var res newModelResponse
	err := c.Post("/api/v2/train/ndb-retrain").Json(body).Do(&res)
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

type saveRequest struct {
	Override  bool   `json:"override"`
	ModelName string `json:"model_name"`
}

type saveReponse struct {
	NewModelId uuid.UUID `json:"new_model_id"`
}

func (c *NdbClient) Save(newModelName string) (*NdbClient, error) {
	body := saveRequest{Override: false, ModelName: newModelName}

	var res wrappedData[saveReponse]
	err := c.Post(fmt.Sprintf("/%v/save", c.deploymentId())).Json(body).Do(&res)
	if err != nil {
		return nil, err
	}

	return &NdbClient{
		ModelClient{
			BaseClient: c.BaseClient,
			modelId:    res.Data.NewModelId,
		},
	}, nil
}

type signedUrlResponse struct {
	SignedUrl string `json:"signed_url"`
}

func (c *NdbClient) GetSignedUrl(source, provider string) (string, error) {
	var res wrappedData[signedUrlResponse]
	err := c.Get(fmt.Sprintf("/%v/get-signed-url", c.deploymentId())).Param("source", source).Param("provider", provider).Do(&res)
	return res.Data.SignedUrl, err
}

func (c *NdbClient) ClientForDeployment(name string) *NdbClient {
	return &NdbClient{ModelClient{
		BaseClient:     c.BaseClient,
		modelId:        c.modelId,
		deploymentName: &name,
	}}
}
