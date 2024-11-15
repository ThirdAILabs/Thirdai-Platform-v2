package client

import (
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

func (c *ModelClient) getStatus(job string) (bool, error) {
	u, err := url.JoinPath(c.baseUrl, fmt.Sprintf("/api/v2/%v/status", job))
	if err != nil {
		return false, fmt.Errorf("error formatting url: %w", err)
	}

	res, err := get[services.StatusResponse](u, map[string]string{"model_id": c.modelId}, c.authToken)
	if err != nil {
		return false, err
	}

	if res.Status == "failed" || res.Status == "stopped" {
		return false, fmt.Errorf("%v has status: %v", job, res.Status)
	}

	return res.Status == "complete", nil
}

func (c *ModelClient) awaitJob(job string, timeout time.Duration) error {
	check := time.Tick(time.Second)
	stop := time.Tick(timeout)
	for {
		select {
		case <-check:
			done, err := c.getStatus(job)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		case <-stop:
			return fmt.Errorf("timeout reached before %v job completed", job)
		}
	}
}

func (c *ModelClient) TrainComplete() (bool, error) {
	return c.getStatus("train")
}

func (c *ModelClient) DeployComplete() (bool, error) {
	return c.getStatus("deploy")
}

func (c *ModelClient) AwaitTrain(timeout time.Duration) error {
	return c.awaitJob("train", timeout)
}

func (c *ModelClient) AwaitDeploy(timeout time.Duration) error {
	return c.awaitJob("deploy", timeout)
}

func (c *ModelClient) Deploy() error {
	body := []byte(fmt.Sprintf(`{"model_id": "%v"}`, c.modelId))

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
