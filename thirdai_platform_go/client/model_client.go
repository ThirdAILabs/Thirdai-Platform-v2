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
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/%v/%v/status", job, c.modelId))
	if err != nil {
		return services.StatusResponse{}, fmt.Errorf("error formatting url: %w", err)
	}

	return get[services.StatusResponse](u, c.authToken)
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

type Logs struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

func (c *ModelClient) getLogs(job string) ([]Logs, error) {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/%v/%v/logs", job, c.modelId))
	if err != nil {
		return nil, fmt.Errorf("error formatting url: %w", err)
	}

	return get[[]Logs](u, c.authToken)
}

func (c *ModelClient) TrainLogs() ([]Logs, error) {
	return c.getLogs("train")
}

func (c *ModelClient) DeployLogs() ([]Logs, error) {
	return c.getLogs("deploy")
}

func (c *ModelClient) Deploy(autoscaling bool) error {
	return c.DeployWithName(autoscaling, "")
}

type deployParams struct {
	Autoscaling    bool   `json:"autoscaling_enabled"`
	DeploymentName string `json:"deployment_name"`
}

func (c *ModelClient) DeployWithName(autoscaling bool, name string) error {
	params := deployParams{
		Autoscaling:    autoscaling,
		DeploymentName: name,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error encoding request body: %w", err)
	}

	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/deploy/%v", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	_, err = post[noBody](u, body, c.authToken)
	return err
}

func (c *ModelClient) Undeploy() error {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/deploy/%v", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	return deleteReq(u, c.authToken)
}

func (c *ModelClient) Delete() error {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/model/%v", c.modelId))
	if err != nil {
		return fmt.Errorf("error formatting url: %w", err)
	}

	return deleteReq(u, c.authToken)
}
