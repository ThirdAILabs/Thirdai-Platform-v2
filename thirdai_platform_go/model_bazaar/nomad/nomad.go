package nomad

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// This will load the given templates into the embed FS so that they are bunddled
// into the compiled binary.

//go:embed jobs/*
var jobTemplates embed.FS

type NomadClient interface {
	StartJob(job Job) error

	StopJob(jobName string) error

	TotalCpuUsage() (int, error)
}

type NomadHttpClient struct {
	addr      string
	token     string
	templates *template.Template
}

func NewHttpClient() NomadClient {
	funcs := template.FuncMap{
		"isLocal": func(d Driver) bool {
			return d.DriverType() == "local"
		},
		"isDocker": func(d Driver) bool {
			return d.DriverType() == "docker"
		},
		"replaceHyphen": func(s string) string {
			return strings.Replace(s, "-", "_", -1)
		},
	}

	tmpl, err := template.New("job_templates").Funcs(funcs).ParseFS(jobTemplates, "jobs/*")
	if err != nil {
		log.Panicf("error parsing job templates: %v", err)
	}

	return &NomadHttpClient{templates: tmpl}
}

func (c *NomadHttpClient) parseJob(job Job) (interface{}, error) {
	content := strings.Builder{}
	err := c.templates.ExecuteTemplate(&content, job.TemplateName(), job)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %v", err)
	}

	payload := map[string]interface{}{"JobHCL": content.String(), "Canonicalize": true}

	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("error encoding job payload: %w", err)
	}

	url, err := url.JoinPath(c.addr, "v1/jobs/parse")
	if err != nil {
		return nil, fmt.Errorf("error formatting job parse url: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending parse request to nomad: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("parse nomad job returned status %d", res.StatusCode)
	}

	defer res.Body.Close()
	var jobDef interface{}
	err = json.NewDecoder(res.Body).Decode(&job)
	if err != nil {
		return nil, fmt.Errorf("error parsing nomad response from parse request: %v", err)
	}

	return jobDef, nil
}

func (c *NomadHttpClient) submitJob(jobDef interface{}) error {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(map[string]interface{}{"Job": jobDef})
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.addr, "v1/jobs")
	if err != nil {
		return fmt.Errorf("error formatting job submit url: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("error creating new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error submitting job to nomad: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("submit nomad job returned status %d", res.StatusCode)
	}

	return nil
}

func (c *NomadHttpClient) StartJob(job Job) error {
	jobDef, err := c.parseJob(job)
	if err != nil {
		return fmt.Errorf("error starting nomad job: %w", err)
	}

	err = c.submitJob(jobDef)
	if err != nil {
		return fmt.Errorf("error starting nomad job: %w", err)
	}

	return nil
}

func (c *NomadHttpClient) StopJob(jobName string) error {
	url, err := url.JoinPath(c.addr, fmt.Sprintf("v1/job/%v", jobName))
	if err != nil {
		return fmt.Errorf("error formatting stop job url: %v", err)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting nomad job: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("delete nomad job returned status %d", res.StatusCode)
	}

	return nil
}

type nomadAllocation struct {
	ClientStatus       string
	AllocatedResources struct {
		Tasks map[string]struct {
			Cpu struct {
				CpuShares int
			}
		}
	}
}

func (c *NomadHttpClient) TotalCpuUsage() (int, error) {
	url, err := url.JoinPath(c.addr, "v1/allocations")
	if err != nil {
		return 0, fmt.Errorf("error formatting allocations url: %w", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating new request: %w", err)
	}
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error getting nomad job allocations: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("list nomad allocations returned status %d", res.StatusCode)
	}

	var allocations []nomadAllocation
	err = json.NewDecoder(res.Body).Decode(&allocations)
	if err != nil {
		return 0, fmt.Errorf("error decoding response from nomad: %w", err)
	}

	totalUsage := 0
	for _, alloc := range allocations {
		for _, task := range alloc.AllocatedResources.Tasks {
			totalUsage += (task.Cpu.CpuShares)
		}
	}

	return totalUsage, nil
}
