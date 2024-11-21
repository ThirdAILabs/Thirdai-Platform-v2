package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"thirdai_platform/model_bazaar/services"
)

type NlpTokenClient struct {
	ModelClient
}

type nlpPredictParams struct {
	Text string `json:"text"`
	Topk int    `json:"top_k"`
}

type NlpTokenPredictions struct {
	Tokens        []string   `json:"tokens"`
	PredictedTags [][]string `json:"predicted_tags"`
}

type nlpPredictResults[T any] struct {
	Data struct {
		PredictionResults T `json:"prediction_results"`
	} `json:"data"`
}

func (c *NlpTokenClient) Predict(text string, topk int) (NlpTokenPredictions, error) {
	params := nlpPredictParams{Text: text, Topk: topk}
	body, err := json.Marshal(params)
	if err != nil {
		return NlpTokenPredictions{}, fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/predict", c.modelId))
	if err != nil {
		return NlpTokenPredictions{}, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[nlpPredictResults[NlpTokenPredictions]](u, body, c.authToken)
	if err != nil {
		return NlpTokenPredictions{}, err
	}

	return res.Data.PredictionResults, nil
}

func (c *NlpTokenClient) AddSample(tokens, tags []string) error {
	params := map[string][]string{"tokens": tokens, "tags": tags}

	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/insert_sample", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, c.authToken)
	return err
}

type newLabel struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

func (c *NlpTokenClient) AddLabel(label, description string, examples []string) error {
	params := map[string][]newLabel{
		"tags": {{Name: label, Description: description, Examples: examples}},
	}

	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/add_labels", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, c.authToken)
	return err
}

func (c *NlpTokenClient) Retrain(name string) (*NlpTokenClient, error) {
	params := services.NlpTokenRetrainRequest{
		ModelName:   name,
		BaseModelId: c.modelId,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/train/nlp-token-retrain")
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[map[string]string](u, body, c.authToken)
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

type NlpTextClient struct {
	ModelClient
}

type PredictedClass struct {
	Class string  `json:"class"`
	Score float32 `json:"score"`
}

type NlpTextPredictions struct {
	PredictedClasses []PredictedClass `json:"predicted_classes"`
}

func (c *NlpTextClient) Predict(text string, topk int) (NlpTextPredictions, error) {
	params := nlpPredictParams{Text: text, Topk: topk}
	body, err := json.Marshal(params)
	if err != nil {
		return NlpTextPredictions{}, fmt.Errorf("error encoding request params: %v", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/%v/predict", c.modelId))
	if err != nil {
		return NlpTextPredictions{}, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := post[nlpPredictResults[NlpTextPredictions]](u, body, c.authToken)
	if err != nil {
		return NlpTextPredictions{}, err
	}

	return res.Data.PredictionResults, nil
}
