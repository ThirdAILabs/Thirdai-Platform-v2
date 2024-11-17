package client

import (
	"encoding/json"
	"fmt"
	"net/url"
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
