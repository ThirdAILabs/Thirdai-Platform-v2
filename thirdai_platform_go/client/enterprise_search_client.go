package client

import "fmt"

type EnterpriseSearchClient struct {
	ModelClient
}

type PiiEntity struct {
	Token string `json:"token"`
	Label string `json:"label"`
}

type EnterpriseSearchResults struct {
	QueryText   string            `json:"query_text"`
	References  []NdbSearchResult `json:"references"`
	PiiEntities []PiiEntity       `json:"pii_entities"`
}

type enterpriseSearchResultsWrapped struct {
	Data EnterpriseSearchResults `json:"data"`
}

func (c *EnterpriseSearchClient) Search(query string, topk int) (EnterpriseSearchResults, error) {
	body := ndbSearchParams{Query: query, Topk: topk}

	var res enterpriseSearchResultsWrapped
	err := c.Post(fmt.Sprintf("/%v/search", c.deploymentId())).Json(body).Do(&res)
	if err != nil {
		return EnterpriseSearchResults{}, err
	}

	return res.Data, nil
}

type unredactParams struct {
	Text        string      `json:"text"`
	PiiEntities []PiiEntity `json:"pii_entities"`
}

type unredactResults struct {
	Data struct {
		UnredactedText string `json:"unredacted_text"`
	} `json:"data"`
}

func (c *EnterpriseSearchClient) Unredact(text string, piiEntities []PiiEntity) (string, error) {
	body := unredactParams{Text: text, PiiEntities: piiEntities}

	var res unredactResults
	err := c.Post(fmt.Sprintf("/%v/unredact", c.deploymentId())).Json(body).Do(&res)
	if err != nil {
		return "", err
	}

	return res.Data.UnredactedText, nil
}
