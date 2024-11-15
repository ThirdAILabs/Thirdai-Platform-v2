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

	res, err := post[nlpPredictResults[NlpTokenPredictions]](u, body, nil, c.authToken)
	if err != nil {
		return NlpTokenPredictions{}, err
	}

	return res.Data.PredictionResults, nil
}

type NlpTextClient struct {
	ModelClient
}

type nlpTextPredictionsRaw struct {
	PredictedClasses [][]any `json:"predicted_classes"`
}

type PredictedClass struct {
	Class string
	Score float32
}

type NlpTextPredictions struct {
	PredictedClasses []PredictedClass
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

	res, err := post[nlpPredictResults[nlpTextPredictionsRaw]](u, body, nil, c.authToken)
	if err != nil {
		return NlpTextPredictions{}, err
	}

	results := NlpTextPredictions{
		PredictedClasses: make([]PredictedClass, 0, len(res.Data.PredictionResults.PredictedClasses)),
	}

	for _, pred := range res.Data.PredictionResults.PredictedClasses {
		if len(pred) != 2 {
			return results, fmt.Errorf("invalid results returned from predict")
		}
		class, cok := pred[0].(string)
		score, sok := pred[1].(float32)

		if !cok || !sok {
			return results, fmt.Errorf("invalid results returned from predict")
		}

		results.PredictedClasses = append(results.PredictedClasses, PredictedClass{Class: class, Score: score})
	}

	return results, nil
}
