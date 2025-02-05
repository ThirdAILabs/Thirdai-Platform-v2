package kubernetes

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1" // For Job manifests.
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"thirdai_platform/model_bazaar/orchestrator"
)

//go:embed jobs/*/*
var jobTemplates embed.FS

// KubernetesClient implements orchestrator.Client using client-go.
type KubernetesClient struct {
	namespace string
	templates *template.Template
	clientset *kubernetes.Clientset
}

// NewKubernetesClient creates a new KubernetesClient.
// If kubeconfigPath is non-empty, that configuration is used; otherwise, the in-cluster config is used.
func NewKubernetesClient(endpoint string, token string, namespace string, kubeconfigPath string) orchestrator.Client {
	// Prepare template helper functions.
	funcs := template.FuncMap{
		"replaceHyphen": func(s string) string {
			return strings.Replace(s, "-", "_", -1)
		},
	}
	// Parse all files in jobs/*/*.
	tmpl, err := template.New("job_templates").Funcs(funcs).ParseFS(jobTemplates, "jobs/*/*")
	if err != nil {
		log.Panicf("error parsing job templates: %v", err)
	}

	// Build client-go config.
	var config *rest.Config
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			log.Panicf("error building kubeconfig: %v", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Panicf("error getting in-cluster config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicf("error creating kubernetes clientset: %v", err)
	}

	slog.Info("creating kubernetes client", "namespace", namespace)
	return &KubernetesClient{
		namespace: namespace,
		templates: tmpl,
		clientset: clientset,
	}
}

// StartJob renders and creates all manifests (deployment, service, ingress, job) in the subdirectory
// specified by job.TemplateName().
func (c *KubernetesClient) StartJob(job orchestrator.Job) error {
	slog.Info("starting kubernetes job", "job_name", job.GetJobName(), "template", job.TemplateName())

	// Compute the subdirectory for the job's manifests.
	subDir := fmt.Sprintf("jobs/%s", job.TemplateName())
	entries, err := fs.ReadDir(jobTemplates, subDir)
	if err != nil {
		slog.Error("error reading job templates directory", "directory", subDir, "error", err)
		return fmt.Errorf("error reading job templates directory %s: %w", subDir, err)
	}

	ctx := context.TODO()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		// Only process recognized manifest files.
		if !(strings.HasSuffix(fileName, "deployment.yaml") ||
			strings.HasSuffix(fileName, "service.yaml") ||
			strings.HasSuffix(fileName, "ingress.yaml") ||
			strings.HasSuffix(fileName, "job.yaml")) {
			slog.Info("skipping unrecognized file", "file", fileName)
			continue
		}

		templateName := filepath.Join(job.TemplateName(), fileName)
		var buf strings.Builder
		if err := c.templates.ExecuteTemplate(&buf, templateName, job); err != nil {
			slog.Error("error rendering job template", "template", templateName, "error", err)
			return fmt.Errorf("error rendering job template %s: %w", templateName, err)
		}
		rendered := buf.String()

		var createErr error
		switch {
		case strings.HasSuffix(fileName, "deployment.yaml"):
			var deployment appsv1.Deployment
			if err := yaml.Unmarshal([]byte(rendered), &deployment); err != nil {
				return fmt.Errorf("error unmarshaling deployment YAML %s: %w", templateName, err)
			}
			_, createErr = c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, &deployment, metav1.CreateOptions{})
		case strings.HasSuffix(fileName, "service.yaml"):
			var service corev1.Service
			if err := yaml.Unmarshal([]byte(rendered), &service); err != nil {
				return fmt.Errorf("error unmarshaling service YAML %s: %w", templateName, err)
			}
			_, createErr = c.clientset.CoreV1().Services(c.namespace).Create(ctx, &service, metav1.CreateOptions{})
		case strings.HasSuffix(fileName, "ingress.yaml"):
			var ingress networkingv1.Ingress
			if err := yaml.Unmarshal([]byte(rendered), &ingress); err != nil {
				return fmt.Errorf("error unmarshaling ingress YAML %s: %w", templateName, err)
			}
			_, createErr = c.clientset.NetworkingV1().Ingresses(c.namespace).Create(ctx, &ingress, metav1.CreateOptions{})
		case strings.HasSuffix(fileName, "job.yaml"):
			var jobObj batchv1.Job
			if err := yaml.Unmarshal([]byte(rendered), &jobObj); err != nil {
				return fmt.Errorf("error unmarshaling job YAML %s: %w", templateName, err)
			}
			_, createErr = c.clientset.BatchV1().Jobs(c.namespace).Create(ctx, &jobObj, metav1.CreateOptions{})
		}
		if createErr != nil {
			slog.Error("error creating resource", "template", templateName, "error", createErr)
			return fmt.Errorf("error creating resource %s: %w", templateName, createErr)
		}
		slog.Info("resource created successfully", "template", templateName)
	}

	slog.Info("all resources for job started successfully", "job_name", job.GetJobName())
	return nil
}

// StopJob deletes the Deployment, Service, and Ingress resources for a given jobName.
func (c *KubernetesClient) StopJob(jobName string) error {
	slog.Info("stopping kubernetes job resources", "job_name", jobName)
	var errs []string
	ctx := context.TODO()

	// Delete Deployment.
	if err := c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		slog.Error("error stopping deployment", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Sprintf("deployment: %v", err))
	}

	// Delete Service.
	if err := c.clientset.CoreV1().Services(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		slog.Error("error stopping service", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Sprintf("service: %v", err))
	}

	// Delete Ingress.
	if err := c.clientset.NetworkingV1().Ingresses(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		slog.Error("error stopping ingress", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Sprintf("ingress: %v", err))
	}

	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("error stopping job resources for %s: %s", jobName, strings.Join(errs, "; ")))
	}
	slog.Info("kubernetes job resources stopped successfully", "job_name", jobName)
	return nil
}

// JobInfo retrieves a simple status summary of the Deployment with the given jobName.
func (c *KubernetesClient) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	ctx := context.TODO()
	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		slog.Error("error getting deployment", "job_name", jobName, "error", err)
		return orchestrator.JobInfo{}, fmt.Errorf("error getting deployment %s: %w", jobName, err)
	}

	status := "Unknown"
	if deployment.Status.ReadyReplicas > 0 {
		status = "Active"
	} else if deployment.Status.UnavailableReplicas > 0 {
		status = "Unavailable"
	}

	info := orchestrator.JobInfo{
		Name:   deployment.Name,
		Status: status,
	}
	slog.Info("job info retrieved", "job_name", jobName, "status", status)
	return info, nil
}

// JobLogs retrieves logs from Pods that belong to the given job.
// This example assumes that Pods are labeled with "app" equal to the jobName.
func (c *KubernetesClient) JobLogs(jobName string) ([]orchestrator.JobLog, error) {
	ctx := context.TODO()
	podList, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", jobName),
	})
	if err != nil {
		slog.Error("error listing pods for job", "job_name", jobName, "error", err)
		return nil, fmt.Errorf("error listing pods for job %s: %w", jobName, err)
	}

	var logs []orchestrator.JobLog
	for _, pod := range podList.Items {
		podLog, err := c.getPodLogs(pod.Name)
		if err != nil {
			slog.Error("error getting logs for pod", "pod", pod.Name, "error", err)
			return nil, err
		}
		logs = append(logs, orchestrator.JobLog{
			Stdout: podLog,
			Stderr: "",
		})
	}
	slog.Info("job logs retrieved", "job_name", jobName)
	return logs, nil
}

// getPodLogs retrieves the logs for a given pod using the client-go PodLogs API.
func (c *KubernetesClient) getPodLogs(podName string) (string, error) {
	ctx := context.TODO()
	podLogOpts := corev1.PodLogOptions{}
	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, &podLogOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error opening log stream for pod %s: %w", podName, err)
	}
	defer stream.Close()

	var builder strings.Builder
	_, err = io.Copy(&builder, stream)
	if err != nil {
		return "", fmt.Errorf("error reading log stream for pod %s: %w", podName, err)
	}
	return builder.String(), nil
}

// ListServices returns information about services and their endpoints in the namespace.
func (c *KubernetesClient) ListServices() ([]orchestrator.ServiceInfo, error) {
	ctx := context.TODO()
	svcList, err := c.clientset.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("error listing services", "error", err)
		return nil, fmt.Errorf("error listing services: %w", err)
	}

	var services []orchestrator.ServiceInfo
	for _, svc := range svcList.Items {
		endpoints, err := c.clientset.CoreV1().Endpoints(c.namespace).Get(ctx, svc.Name, metav1.GetOptions{})
		if err != nil {
			slog.Error("error getting endpoints for service", "service", svc.Name, "error", err)
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
						Port:    int(port.Port),
					}
					if addr.TargetRef != nil {
						allocation.AllocID = addr.TargetRef.Name
					}
					allocations = append(allocations, allocation)
				}
			}
		}

		services = append(services, orchestrator.ServiceInfo{
			Name:        svc.Name,
			Allocations: allocations,
		})
	}
	slog.Info("services listed", "count", len(services))
	return services, nil
}

// TotalCpuUsage calculates the total CPU usage (in millicores) of running pods in the namespace.
func (c *KubernetesClient) TotalCpuUsage() (int, error) {
	ctx := context.TODO()
	podList, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("error listing pods", "error", err)
		return 0, fmt.Errorf("error listing pods: %w", err)
	}

	totalMillicores := 0
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if cpu, ok := container.Resources.Requests["cpu"]; ok {
				m, err := parseCPUQuantity(cpu.String())
				if err != nil {
					slog.Error("error parsing cpu quantity", "quantity", cpu.String(), "error", err)
					continue
				}
				totalMillicores += m
			}
		}
	}
	slog.Info("total CPU usage calculated", "totalMillicores", totalMillicores)
	return totalMillicores, nil
}

// parseCPUQuantity converts a Kubernetes CPU quantity string (e.g., "250m" or "1") into an int value in millicores.
func parseCPUQuantity(q string) (int, error) {
	if strings.HasSuffix(q, "m") {
		trimmed := strings.TrimSuffix(q, "m")
		m, err := strconv.Atoi(trimmed)
		if err != nil {
			return 0, err
		}
		return m, nil
	}
	f, err := strconv.ParseFloat(q, 64)
	if err != nil {
		return 0, err
	}
	return int(f * 1000), nil
}
