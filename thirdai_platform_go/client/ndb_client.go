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
	Data struct {
		References []NdbSearchResult `json:"references"`
	} `json:"data"`
}

func (c *NdbClient) Search(query string, topk int) ([]NdbSearchResult, error) {
	params := ndbSearchParams{Query: query, Topk: topk}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/search", c.modelId))
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[ndbSearchResults](u, body, nil, c.authToken)
	if err != nil {
		return nil, err
	}

	return res.Data.References, nil
}

type insertParams struct {
	Documents []config.FileInfo `json:"documents"`
}

func (c *NdbClient) Insert(files []config.FileInfo) error {
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

	params := insertParams{Documents: updateLocalFilePrefixes(files, "")}
	for i := range params.Documents {
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

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/insert", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	headers := authHeader(c.authToken)
	headers["Content-Type"] = writer.FormDataContentType()

	_, err = postWithHeaders[noBody](u, body.Bytes(), nil, headers)
	return err
}

type deleteParams struct {
	SourceIds []string `json:"source_ids"`
}

func (c *NdbClient) Delete(doc_ids []string) error {
	params := deleteParams{SourceIds: doc_ids}
	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/delete", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, nil, c.authToken)
	return err
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
	params := upvoteParams{TextIdPairs: samples}
	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/upvote", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, nil, c.authToken)
	return err
}

type AssociatePair struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type associateParams struct {
	TextPairs []AssociatePair `json:"text_pairs"`
}

func (c *NdbClient) Associate(samples []AssociatePair) error {
	params := associateParams{TextPairs: samples}
	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/associate", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, nil, c.authToken)
	return err
}

type Source struct {
	Source   string `json:"source"`
	SourceId string `json:"source_id"`
	Version  int    `json:"version"`
}

type sourcesResponse struct {
	Data []Source `json:"data"`
}

func (c *NdbClient) Sources() ([]Source, error) {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/sources", c.modelId))
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := get[sourcesResponse](u, nil, c.authToken)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *NdbClient) Retrain(newModelName string) (*NdbClient, error) {
	params := services.NdbRetrainOptions{
		ModelName:   newModelName,
		BaseModelId: c.modelId,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/train/ndb-retrain")
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
