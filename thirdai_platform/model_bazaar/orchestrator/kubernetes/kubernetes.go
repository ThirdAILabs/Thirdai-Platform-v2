package kubernetes

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"thirdai_platform/model_bazaar/orchestrator"
)

//go:embed jobs/*
var jobTemplates embed.FS

type KubernetesClient struct {
	endpoint  string
	token     string
	namespace string
	templates *template.Template
}

func NewKubernetesClient(endpoint string, token string, namespace string) orchestrator.Client {
	// Here you might define any helper functions for your templates.
	funcs := template.FuncMap{
		"replaceHyphen": func(s string) string {
			return strings.Replace(s, "-", "_", -1)
		},
	}

	tmpl, err := template.New("job_templates").Funcs(funcs).ParseFS(jobTemplates, "jobs/*")
	if err != nil {
		log.Panicf("error parsing job templates: %v", err)
	}

	slog.Info("creating kubernetes client", "endpoint", endpoint, "namespace", namespace)
	return &KubernetesClient{
		endpoint:  endpoint,
		token:     token,
		namespace: namespace,
		templates: tmpl,
	}
}

// request is a helper that constructs a full URL from the Kubernetes API endpoint,
// sets common headers (including the Bearer token if one is provided),
// performs the request, and (if result is non-nil) decodes the JSON response.
func (c *KubernetesClient) request(method, apiPath string, body io.Reader, result interface{}) error {
	fullURL, err := url.JoinPath(c.endpoint, apiPath)
	if err != nil {
		return fmt.Errorf("error formatting URL for kubernetes API path %v: %w", apiPath, err)
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return fmt.Errorf("error creating %v request for kubernetes API path %v: %w", method, apiPath, err)
	}
	req.Header.Add("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Add("Authorization", "Bearer "+c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending %v request to kubernetes API path %v: %w", method, apiPath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		slog.Error("kubernetes API returned error", "method", method, "apiPath", apiPath, "code", resp.StatusCode, "response", string(data))
		return fmt.Errorf("%v request to kubernetes API path %v returned status %d", method, apiPath, resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("error decoding %v response from kubernetes API path %v: %w", method, apiPath, err)
		}
	}
	return nil
}

func (c *KubernetesClient) StartJob(job orchestrator.Job) error {
	slog.Info("starting kubernetes job", "job_name", job.GetJobName(), "template", job.TemplateName())

	// Render the job manifest (assumed to be in JSON or YAML)
	var buf strings.Builder
	if err := c.templates.ExecuteTemplate(&buf, job.TemplateName(), job); err != nil {
		slog.Error("error rendering job template", "job_name", job.GetJobName(), "error", err)
		return fmt.Errorf("error rendering job template: %w", err)
	}

	// Post the rendered manifest to the Batch API.
	apiPath := fmt.Sprintf("apis/batch/v1/namespaces/%s/jobs", c.namespace)
	body := bytes.NewBufferString(buf.String())
	if err := c.request("POST", apiPath, body, nil); err != nil {
		slog.Error("error submitting kubernetes job", "job_name", job.GetJobName(), "error", err)
		return fmt.Errorf("error starting kubernetes job: %w", err)
	}

	slog.Info("kubernetes job started successfully", "job_name", job.GetJobName())
	return nil
}

// StopJob deletes the Kubernetes Job resource.
func (c *KubernetesClient) StopJob(jobName string) error {
	slog.Info("stopping kubernetes job", "job_name", jobName)
	apiPath := fmt.Sprintf("apis/batch/v1/namespaces/%s/jobs/%s", c.namespace, jobName)
	if err := c.request("DELETE", apiPath, nil, nil); err != nil {
		slog.Error("error stopping kubernetes job", "job_name", jobName, "error", err)
		return fmt.Errorf("error stopping kubernetes job %s: %w", jobName, err)
	}
	slog.Info("kubernetes job stopped successfully", "job_name", jobName)
	return nil
}

// JobInfo fetches the Kubernetes Job resource and returns a simple status summary.
func (c *KubernetesClient) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	slog.Info("retrieving kubernetes job info", "job_name", jobName)
	apiPath := fmt.Sprintf("apis/batch/v1/namespaces/%s/jobs/%s", c.namespace, jobName)
	var jobResp struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			Active    int `json:"active"`
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
		} `json:"status"`
	}
	if err := c.request("GET", apiPath, nil, &jobResp); err != nil {
		slog.Error("error getting kubernetes job info", "job_name", jobName, "error", err)
		return orchestrator.JobInfo{}, fmt.Errorf("error getting info for kubernetes job %s: %w", jobName, err)
	}

	status := "Unknown"
	if jobResp.Status.Active > 0 {
		status = "Active"
	} else if jobResp.Status.Succeeded > 0 {
		status = "Succeeded"
	} else if jobResp.Status.Failed > 0 {
		status = "Failed"
	}

	info := orchestrator.JobInfo{
		Name:   jobResp.Metadata.Name,
		Status: status,
	}

	slog.Info("kubernetes job info retrieved", "job_name", jobName, "status", status)
	return info, nil
}

// getPodLogs retrieves the logs for a given Pod.
func (c *KubernetesClient) getPodLogs(podName string) (string, error) {
	logPath := fmt.Sprintf("api/v1/namespaces/%s/pods/%s/log", c.namespace, podName)
	fullURL, err := url.JoinPath(c.endpoint, logPath)
	if err != nil {
		return "", fmt.Errorf("error constructing log URL for pod %s: %w", podName, err)
	}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for pod %s logs: %w", podName, err)
	}
	if c.token != "" {
		req.Header.Add("Authorization", "Bearer "+c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error retrieving logs for pod %s: %w", podName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error retrieving logs for pod %s: status %d, response: %s", podName, resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading logs for pod %s: %w", podName, err)
	}
	return string(data), nil
}

// JobLogs finds the Pods created by the Job (by selecting on label "job-name")
// and returns their logs.
func (c *KubernetesClient) JobLogs(jobName string) ([]orchestrator.JobLog, error) {
	slog.Info("retrieving kubernetes job logs", "job_name", jobName)
	apiPath := fmt.Sprintf("api/v1/namespaces/%s/pods?labelSelector=job-name=%s", c.namespace, jobName)
	var podList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := c.request("GET", apiPath, nil, &podList); err != nil {
		slog.Error("error listing pods for job", "job_name", jobName, "error", err)
		return nil, fmt.Errorf("error listing pods for job %s: %w", jobName, err)
	}

	var logs []orchestrator.JobLog
	for _, pod := range podList.Items {
		podLog, err := c.getPodLogs(pod.Metadata.Name)
		if err != nil {
			slog.Error("error getting logs for pod", "pod", pod.Metadata.Name, "error", err)
			return nil, err
		}
		logs = append(logs, orchestrator.JobLog{
			Stdout: podLog,
			Stderr: "", // Kubernetes does not separate stdout/stderr
		})
	}
	slog.Info("kubernetes job logs retrieved", "job_name", jobName)
	return logs, nil
}

// ListServices returns information about services in the namespace.
// For each Service it fetches its Endpoints and creates a list of allocations.
func (c *KubernetesClient) ListServices() ([]orchestrator.ServiceInfo, error) {
	slog.Info("listing kubernetes services", "namespace", c.namespace)
	apiPath := fmt.Sprintf("api/v1/namespaces/%s/services", c.namespace)
	var svcList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := c.request("GET", apiPath, nil, &svcList); err != nil {
		slog.Error("error listing services", "error", err)
		return nil, fmt.Errorf("error listing services: %w", err)
	}

	var services []orchestrator.ServiceInfo
	for _, svc := range svcList.Items {
		// Get endpoints for the service.
		endpointsPath := fmt.Sprintf("api/v1/namespaces/%s/endpoints/%s", c.namespace, svc.Metadata.Name)
		var endpoints struct {
			Subsets []struct {
				Addresses []struct {
					IP        string `json:"ip"`
					TargetRef *struct {
						Name string `json:"name"`
					} `json:"targetRef,omitempty"`
				} `json:"addresses"`
				Ports []struct {
					Port int `json:"port"`
				} `json:"ports"`
			} `json:"subsets"`
		}
		if err := c.request("GET", endpointsPath, nil, &endpoints); err != nil {
			// If endpoints are not available, skip this service.
			slog.Error("error getting endpoints for service", "service", svc.Metadata.Name, "error", err)
			continue
		}

		var allocations []orchestrator.ServiceAllocation
		for _, subset := range endpoints.Subsets {
			for _, addr := range subset.Addresses {
				for _, port := range subset.Ports {
					allocation := orchestrator.ServiceAllocation{
						Address: addr.IP,
						AllocID: "",
						NodeID:  "",
						Port:    port.Port,
					}
					if addr.TargetRef != nil {
						allocation.AllocID = addr.TargetRef.Name
					}
					allocations = append(allocations, allocation)
				}
			}
		}
		services = append(services, orchestrator.ServiceInfo{
			Name:        svc.Metadata.Name,
			Allocations: allocations,
		})
	}
	slog.Info("kubernetes services listed", "count", len(services))
	return services, nil
}

func (c *KubernetesClient) TotalCpuUsage() (int, error) {
	slog.Info("calculating total CPU usage from running pods")
	apiPath := fmt.Sprintf("api/v1/namespaces/%s/pods", c.namespace)
	var podList struct {
		Items []struct {
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
			Spec struct {
				Containers []struct {
					Resources struct {
						Requests map[string]string `json:"requests"`
					} `json:"resources"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := c.request("GET", apiPath, nil, &podList); err != nil {
		slog.Error("error listing pods", "error", err)
		return 0, fmt.Errorf("error listing pods: %w", err)
	}

	totalMillicores := 0
	for _, pod := range podList.Items {
		if pod.Status.Phase != "Running" {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if cpu, ok := container.Resources.Requests["cpu"]; ok {
				m, err := parseCPUQuantity(cpu)
				if err != nil {
					slog.Error("error parsing cpu quantity", "quantity", cpu, "error", err)
					continue
				}
				totalMillicores += m
			}
		}
	}
	slog.Info("total CPU usage calculated", "totalMillicores", totalMillicores)
	return totalMillicores, nil
}

// parseCPUQuantity converts a Kubernetes CPU quantity (for example, "250m" or "1")
// into an integer value in millicores.
func parseCPUQuantity(q string) (int, error) {
	if strings.HasSuffix(q, "m") {
		trimmed := strings.TrimSuffix(q, "m")
		m, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return m, nil
	}
	// Assume it is specified in CPU cores.
	f, err := strconv.ParseFloat(q, 64)
	if err != nil {
		return 0, err
	}
	return int(f * 1000), nil
}
