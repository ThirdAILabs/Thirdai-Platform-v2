package client

import (
	"fmt"
	"io"
	"os"
	"thirdai_platform/model_bazaar/services"
	"time"

	"github.com/google/uuid"
)

type ModelClient struct {
	BaseClient
	modelId        uuid.UUID
	deploymentName *string
}

func NewModelClient(baseUrl string, authToken string, modelId uuid.UUID) ModelClient {
	return ModelClient{BaseClient: BaseClient{baseUrl: baseUrl, authToken: authToken}, modelId: modelId}
}

func (c *ModelClient) GetModelID() uuid.UUID {
	return c.modelId
}

func (c *ModelClient) deploymentId() string {
	if c.deploymentName != nil {
		return *c.deploymentName
	}
	return c.modelId.String()
}

func (c *ModelClient) DeploymentHealthy() bool {
	err := c.Get(fmt.Sprintf("/%v/health", c.deploymentId())).Do(nil)
	return err == nil
}

func (c *ModelClient) getStatus(job string) (services.StatusResponse, error) {
	var res services.StatusResponse
	err := c.Get(fmt.Sprintf("/api/v2/%v/%v/status", job, c.modelId)).Do(&res)

	return res, err
}

func (c *ModelClient) awaitJob(job string, timeout time.Duration) error {
	check := time.Tick(2 * time.Second)
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
				if job == "train" || c.DeploymentHealthy() {
					return nil
				}
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
	var res []Logs
	err := c.Get(fmt.Sprintf("/api/v2/%v/%v/logs", job, c.modelId)).Do(&res)
	return res, err
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
	body := deployParams{
		Autoscaling:    autoscaling,
		DeploymentName: name,
	}

	return c.Post(fmt.Sprintf("/api/v2/deploy/%v", c.modelId)).Json(body).Do(nil)
}

func (c *ModelClient) Undeploy() error {
	return c.Delete(fmt.Sprintf("/api/v2/deploy/%v", c.modelId)).Do(nil)
}

func (c *ModelClient) DeleteModel() error {
	return c.Delete(fmt.Sprintf("/api/v2/model/%v", c.modelId)).Do(nil)
}

func (c *ModelClient) Download(dstPath string) error {
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	return c.Get(fmt.Sprintf("/api/v2/model/%v/download", c.modelId)).Process(
		func(body io.Reader) error {
			_, err := io.Copy(dst, body)
			return err
		},
	)
}

func (c *ModelClient) GetPermissions() (services.ModelPermissions, error) {
	var res services.ModelPermissions
	err := c.Get(fmt.Sprintf("/api/v2/model/%v/permissions", c.modelId)).Do(&res)
	return res, err
}

type newModelResponse struct {
	ModelId uuid.UUID `json:"model_id"`
}
