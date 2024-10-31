package nomad

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

type NomadClient interface {
	StartJob(jobTemplate string, args interface{}) error

	StopJob(jobName string) error
}

type NomadHttpClient struct {
	addr  string
	token string
}

func NewHttpClient() NomadClient {
	return &NomadHttpClient{}
}

func (c *NomadHttpClient) parseJob(jobTemplate string, args interface{}) (interface{}, error) {
	tmpl, err := template.New("").ParseFiles(jobTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing job template: %v", err)
	}

	content := strings.Builder{}
	err = tmpl.Execute(&content, args)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %v", err)
	}

	payload := map[string]interface{}{"JobHCL": content.String(), "Canonicalize": true}

	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.JoinPath(c.addr, "v1/jobs/parse")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending parse request to nomad: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("parse nomad job returned status %d", res.StatusCode)
	}

	defer res.Body.Close()
	var job interface{}
	err = json.NewDecoder(res.Body).Decode(&job)
	if err != nil {
		return nil, fmt.Errorf("error parsing nomad response from parse request: %v", err)
	}

	return job, nil
}

func (c *NomadHttpClient) submitJob(jobDef interface{}) error {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(map[string]interface{}{"Job": jobDef})
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.addr, "v1/jobs")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error submitting job to nomad: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("submit nomad job returned status %d", res.StatusCode)
	}

	return nil
}

func (c *NomadHttpClient) StartJob(jobTemplate string, args interface{}) error {
	job, err := c.parseJob(jobTemplate, args)
	if err != nil {
		return err
	}

	return c.submitJob(job)
}

func (c *NomadHttpClient) StopJob(jobName string) error {
	url, err := url.JoinPath(c.addr, fmt.Sprintf("v1/job/%v", jobName))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting nomad job: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("delete nomad job returned status %d", res.StatusCode)
	}

	return nil
}
