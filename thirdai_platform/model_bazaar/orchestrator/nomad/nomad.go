package nomad

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"thirdai_platform/model_bazaar/orchestrator"
)

// This will load the given templates into the embed FS so that they are bundled
// into the compiled binary.

//go:embed jobs/*
var jobTemplates embed.FS

type NomadClient struct {
	addr            string
	token           string
	templates       *template.Template
	ingressHostname string
}

func NomadTemplatePath(jobPath string) string {
	return jobPath + ".hcl.tmpl"
}

func NewNomadClient(addr string, token string, ingressHostname string) orchestrator.Client {
	funcs := template.FuncMap{
		"isLocal": func(d orchestrator.Driver) bool {
			return d.DriverType() == "local"
		},
		"isDocker": func(d orchestrator.Driver) bool {
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

	slog.Info("creating nomad http client")
	for _, t := range tmpl.Templates() {
		slog.Info("found job template: " + t.Name())
	}

	return &NomadClient{addr: addr, token: token, templates: tmpl, ingressHostname: ingressHostname}
}

var errNomadReturnedNotFound = errors.New("nomad returned status 404")

func (c *NomadClient) request(method, endpoint string, body io.Reader, result interface{}) error {
	fullEndpoint, err := url.JoinPath(c.addr, endpoint)
	if err != nil {
		return fmt.Errorf("error formatting url for nomad endpoint %v: %w", endpoint, err)
	}

	req, err := http.NewRequest(method, fullEndpoint, body)
	if err != nil {
		return fmt.Errorf("error creating %v request for nomad endpoint %v: %w", method, endpoint, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nomad-Token", c.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending %v request to nomad endpoint %v: %w", method, endpoint, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return errNomadReturnedNotFound
	}
	if res.StatusCode != http.StatusOK {
		data, err := io.ReadAll(res.Body)
		if err == nil {
			slog.Error("nomad returned error", "method", method, "endpoint", endpoint, "code", res.StatusCode, "response", string(data))
		}
		return fmt.Errorf("%v request to nomad endpoint %v returned status %d", method, endpoint, res.StatusCode)
	}

	if result != nil {
		err := json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error parsing %v response from nomad endpoint %v: %w", method, endpoint, err)
		}
	}

	return nil
}

func (c *NomadClient) get(endpoint string, result interface{}) error {
	return c.request("GET", endpoint, nil, result)
}

func (c *NomadClient) post(endpoint string, body io.Reader, result interface{}) error {
	return c.request("POST", endpoint, body, result)
}

func (c *NomadClient) delete(endpoint string) error {
	return c.request("DELETE", endpoint, nil, nil)
}

func (c *NomadClient) parseJob(job orchestrator.Job) (interface{}, error) {
	content := strings.Builder{}
	err := c.templates.ExecuteTemplate(&content, NomadTemplatePath(job.JobTemplatePath()), job)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %v", err)
	}

	payload := map[string]interface{}{"JobHCL": content.String(), "Canonicalize": true}

	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("error encoding job payload: %w", err)
	}

	var jobDef interface{}
	err = c.post("v1/jobs/parse", body, &jobDef)
	if err != nil {
		return nil, err
	}

	return jobDef, nil
}

func (c *NomadClient) submitJob(jobDef interface{}) error {
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(map[string]interface{}{"Job": jobDef})
	if err != nil {
		return fmt.Errorf("error encoding job submit payload: %w", err)
	}

	err = c.post("v1/jobs", body, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *NomadClient) StartJob(job orchestrator.Job) error {

	nomadTemplatePath := NomadTemplatePath(job.JobTemplatePath())
	slog.Info("starting nomad job", "job_name", job.GetJobName(), "template", nomadTemplatePath)

	jobDef, err := c.parseJob(job)
	if err != nil {
		slog.Error("error parsing nomad job", "job_name", job.GetJobName(), "template", nomadTemplatePath, "error", err)
		return fmt.Errorf("error starting nomad job: %w", err)
	}

	err = c.submitJob(jobDef)
	if err != nil {
		slog.Error("error submitting nomad job", "job_name", job.GetJobName(), "template", nomadTemplatePath, "error", err)
		return fmt.Errorf("error starting nomad job: %w", err)
	}

	slog.Info("nomad job started successfully", "job_name", job.GetJobName(), "template", nomadTemplatePath)

	return nil
}

func (c *NomadClient) StopJob(jobName string) error {
	slog.Info("stopping nomad job", "job_name", jobName)

	err := c.delete(fmt.Sprintf("v1/job/%v", jobName))
	if err != nil {
		slog.Error("error stopping nomad job", "job_name", jobName, "error", err)
		return fmt.Errorf("error stopping nomad job %v: %w", jobName, err)
	}

	slog.Info("nomad job stopped successfully", "job_name", jobName)

	return nil
}

func (c *NomadClient) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	slog.Debug("retrieving nomad job info", "job_name", jobName)

	var info orchestrator.JobInfo
	err := c.get(fmt.Sprintf("v1/job/%v", jobName), &info)
	if err != nil {
		if errors.Is(err, errNomadReturnedNotFound) {
			return orchestrator.JobInfo{}, orchestrator.ErrJobNotFound
		}
		slog.Error("error getting nomad job info", "job_name", jobName, "error", err)
		return orchestrator.JobInfo{}, fmt.Errorf("error getting info for nomad job %v: %w", jobName, err)
	}

	slog.Debug("nomad job info retrieved successfully", "job_name", jobName)

	return info, nil
}

type jobAllocation struct {
	ID string
}

func (c *NomadClient) jobAllocations(jobName string) ([]string, error) {
	var allocations []jobAllocation
	err := c.get(fmt.Sprintf("v1/job/%v/allocations", jobName), &allocations)
	if err != nil {
		return nil, fmt.Errorf("error retreiving allocations for nomad job %v: %w", jobName, err)
	}

	allocIds := make([]string, 0, len(allocations))
	for _, alloc := range allocations {
		allocIds = append(allocIds, alloc.ID)
	}

	return allocIds, nil
}

func (c *NomadClient) getLogs(allocId string, logType string) (string, error) {
	url, err := url.JoinPath(c.addr, fmt.Sprintf("v1/client/fs/logs/%v", allocId))
	if err != nil {
		return "", fmt.Errorf("error formatting allocation logs url: %w", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating new request: %w", err)
	}
	req.Header.Add("X-Nomad-Token", c.token)

	params := map[string]string{
		"task": "backend", "type": logType, "origin": "end", "offset": "5000", "plain": "true",
	}
	query := req.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error retrieving nomad logs: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get nomad job logs returned status %d", res.StatusCode)
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading log response: %w", err)
	}

	return string(content), err
}

func (c *NomadClient) JobLogs(jobName string) ([]orchestrator.JobLog, error) {
	slog.Info("retrieving nomad job logs", "job_name", jobName)

	allocations, err := c.jobAllocations(jobName)
	if err != nil {
		slog.Error("error listing job allocations", "job_name", jobName, "error", err)
		return nil, fmt.Errorf("error listing allcoations for job %v: %w", jobName, err)
	}

	logs := make([]orchestrator.JobLog, 0)

	for _, alloc := range allocations {
		stdoutLogs, err := c.getLogs(alloc, "stdout")
		if err != nil {
			slog.Error("error getting stdout logs", "job_name", jobName, "error", err)
			return nil, fmt.Errorf("error getting logs from stdout for job %v: %w", jobName, err)
		}
		stderrLogs, err := c.getLogs(alloc, "stderr")
		if err != nil {
			slog.Error("error getting stderr logs", "job_name", jobName, "error", err)
			return nil, fmt.Errorf("error getting logs from stderr for job %v: %w", jobName, err)
		}

		logs = append(logs, orchestrator.JobLog{Stdout: stdoutLogs, Stderr: stderrLogs})
	}

	slog.Info("nomad job logs retrieved successfully", "job_name", jobName)

	return logs, nil
}

type serviceResponse struct {
	Namespace string
	Services  []struct {
		ServiceName string
	}
}

func (c *NomadClient) listAllServices() ([]string, error) {
	var namespaces []serviceResponse
	err := c.get("v1/services", &namespaces)
	if err != nil {
		return nil, err
	}

	services := make([]string, 0)
	for _, namespace := range namespaces {
		for _, service := range namespace.Services {
			services = append(services, service.ServiceName)
		}
	}

	return services, nil
}

func (c *NomadClient) getServiceAllocations(service string) (orchestrator.ServiceInfo, error) {
	var allocations []orchestrator.ServiceAllocation
	err := c.get(fmt.Sprintf("v1/service/%v", service), &allocations)
	if err != nil {
		return orchestrator.ServiceInfo{}, nil
	}

	return orchestrator.ServiceInfo{Name: service, Allocations: allocations}, nil
}

func (c *NomadClient) ListServices() ([]orchestrator.ServiceInfo, error) {
	serviceNames, err := c.listAllServices()
	if err != nil {
		slog.Error("error listing nomad services", "error", err)
		return nil, fmt.Errorf("error listing nomad services: %w", err)
	}

	serviceInfos := make([]orchestrator.ServiceInfo, 0, len(serviceNames))
	for _, service := range serviceNames {
		info, err := c.getServiceAllocations(service)
		if err != nil {
			slog.Error("error getting info for nomad service", "service", service, "error", err)
			return nil, fmt.Errorf("error getting info for service %v: %w", service, err)
		}
		serviceInfos = append(serviceInfos, info)
	}

	return serviceInfos, nil
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

func (c *NomadClient) TotalCpuUsage() (int, error) {
	slog.Info("getting nomad total cpu usage")

	var allocations []nomadAllocation
	err := c.get("v1/allocations", &allocations)
	if err != nil {
		slog.Error("error getting nomad total cpu usage", "error", err)
		return 0, fmt.Errorf("error getting nomad total cpu usage: %w", err)
	}

	totalUsage := 0
	for _, alloc := range allocations {
		if alloc.ClientStatus == "running" {
			for _, task := range alloc.AllocatedResources.Tasks {
				totalUsage += task.Cpu.CpuShares
			}
		}
	}

	slog.Info("got nomad total cpu usage successfully", "total_cpu_usage", totalUsage)

	return totalUsage, nil
}

func (c *NomadClient) IngressHostname() string {
	return c.ingressHostname
}

func (c *NomadClient) GetName() string {
	return "nomad"
}
