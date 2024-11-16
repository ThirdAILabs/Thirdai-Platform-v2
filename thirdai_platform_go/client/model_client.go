package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"thirdai_platform/model_bazaar/services"
	"time"
)

type ModelClient struct {
	baseUrl   string
	authToken string
	modelId   string
}

func (c *ModelClient) getStatus(job string) (services.StatusResponse, error) {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/%v/status", job))
	if err != nil {
		return services.StatusResponse{}, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := get[services.StatusResponse](u, map[string]string{"model_id": c.modelId}, c.authToken)
	if err != nil {
		return services.StatusResponse{}, err
	}

	return res, nil
}

func (c *ModelClient) awaitJob(job string, timeout time.Duration) error {
	check := time.Tick(time.Second)
	stop := time.Tick(timeout)
	for {
		select {
		case <-check:
			status, err := c.getStatus(job)
			if err != nil {
				return err
			}
			if status.Status == "failed" || status.Status == "stopped" {
				return fmt.Errorf("%v has status: %v", job, status.Status)
			}
			if status.Status == "complete" {
				return nil
			}
		case <-stop:
			return fmt.Errorf("timeout reached before %v job completed", job)
		}
	}
}

func (c *ModelClient) TrainStatus() (services.StatusResponse, error) {
	return c.getStatus("train")
}

func (c *ModelClient) DeployStatus() (services.StatusResponse, error) {
	return c.getStatus("deploy")
}

func (c *ModelClient) AwaitTrain(timeout time.Duration) error {
	return c.awaitJob("train", timeout)
}

func (c *ModelClient) AwaitDeploy(timeout time.Duration) error {
	return c.awaitJob("deploy", timeout)
}

func (c *ModelClient) Deploy(autoscaling bool) error {
	return c.DeployWithName(autoscaling, "")
}

type deployParams struct {
	Autoscaling    bool   `json:"autoscaling_enabled"`
	ModelId        string `json:"model_id"`
	DeploymentName string `json:"deployment_name"`
}

func (c *ModelClient) DeployWithName(autoscaling bool, name string) error {
	params := deployParams{
		Autoscaling:    autoscaling,
		ModelId:        c.modelId,
		DeploymentName: name,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request body: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, "/api/v2/deploy/start")
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, nil, c.authToken)
	return err
}

func (c *ModelClient) Undeploy() error {
	u, err := url.JoinPath(c.baseUrl, "/api/v2/deploy/stop")
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, nil, map[string]string{"model_id": c.modelId}, c.authToken)
	return err
}

func (c *ModelClient) Delete() error {
	u, err := url.JoinPath(c.baseUrl, "/api/v2/model/delete")
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, nil, map[string]string{"model_id": c.modelId}, c.authToken)
	return err
}
