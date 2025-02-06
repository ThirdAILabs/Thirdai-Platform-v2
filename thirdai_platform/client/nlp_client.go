package client

import (
	"fmt"
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
	body := nlpPredictParams{Text: text, Topk: topk}

	var res nlpPredictResults[NlpTokenPredictions]
	err := c.Post(fmt.Sprintf("/%v/predict", c.deploymentId())).Json(body).Do(&res)
	if err != nil {
		return NlpTokenPredictions{}, err
	}

	return res.Data.PredictionResults, nil
}

func (c *NlpTokenClient) AddSample(tokens, tags []string) error {
	body := map[string][]string{"tokens": tokens, "tags": tags}

	return c.Post(fmt.Sprintf("/%v/insert_sample", c.deploymentId())).Json(body).Do(nil)
}

type newLabel struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

func (c *NlpTokenClient) AddLabel(label, description string, examples []string) error {
	body := map[string][]newLabel{
		"tags": {{Name: label, Description: description, Examples: examples}},
	}

	return c.Post(fmt.Sprintf("/%v/add_labels", c.deploymentId())).Json(body).Do(nil)
}

func (c *NlpTokenClient) Retrain(name string) (*NlpTokenClient, error) {
	body := services.NlpTokenRetrainRequest{
		ModelName:   name,
		BaseModelId: c.modelId,
	}

	var res newModelResponse
	err := c.Post("/api/v2/train/nlp-token-retrain").Json(body).Do(&res)
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
	body := nlpPredictParams{Text: text, Topk: topk}

	var res nlpPredictResults[NlpTextPredictions]
	err := c.Post(fmt.Sprintf("/%v/predict", c.deploymentId())).Json(body).Do(&res)
	if err != nil {
		return NlpTextPredictions{}, err
	}

	return res.Data.PredictionResults, nil
}
