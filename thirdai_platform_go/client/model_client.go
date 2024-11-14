package client

import (
	"fmt"
	"net/url"
	"thirdai_platform/model_bazaar/services"
)

type ModelClient struct {
	baseUrl   string
	authToken string
	modelId   string
}

func (c *ModelClient) getStatus(job string) (bool, error) {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/%v/status?model_id=%v", job, c.modelId))
	if err != nil {
		return false, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := get[services.StatusResponse](u, c.authToken)
	if err != nil {
		return false, err
	}

	if res.Status == "failed" || res.Status == "stopped" {
		return false, fmt.Errorf("%v has status: %v", job, res.Status)
	}

	return res.Status == "complete", nil
}

func (c *ModelClient) TrainComplete() (bool, error) {
	return c.getStatus("train")
}

func (c *ModelClient) DeployComplete() (bool, error) {
	return c.getStatus("deploy")
}

func (c *ModelClient) Deploy() error {
	body := []byte(fmt.Sprintf(`{"model_id": "%v"}`, c.modelId))

	u, err := url.JoinPath(c.baseUrl, "/api/v2/deploy/start")
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, c.authToken)
	return err
}

func (c *ModelClient) Undeploy() error {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/deploy/stop?model_id=%v", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, nil, c.authToken)
	return err
}

func (c *ModelClient) Delete() error {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/model/delete?model_id=%v", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, nil, c.authToken)
	return err
}
