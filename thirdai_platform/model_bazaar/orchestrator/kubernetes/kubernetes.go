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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1" // For Job manifests.
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"thirdai_platform/model_bazaar/orchestrator"
)

//go:embed jobs/*/*
var jobTemplates embed.FS

func getNamespace() (string, error) {
	// Read the namespace from the well-known file
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

type KubernetesClient struct {
	namespace string
	templates *template.Template
	clientset *kubernetes.Clientset
}

func NewKubernetesClient() orchestrator.Client {
	tmpl, err := template.New("job_templates").ParseFS(jobTemplates, "jobs/*/*")
	if err != nil {
		log.Panicf("error parsing job templates: %v", err)
	}

	var config *rest.Config

	config, err = rest.InClusterConfig()
	if err != nil {
		log.Panicf("error getting in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicf("error creating kubernetes clientset: %v", err)
	}

	namespace, err := getNamespace()
	if err != nil {
		log.Panicf("error creating retrieving kubernetes namespace: %v", err)
	}

	slog.Info("creating kubernetes client", "namespace", namespace)
	return &KubernetesClient{
		namespace: namespace,
		templates: tmpl,
		clientset: clientset,
	}
}

func (c *KubernetesClient) StartJob(job orchestrator.Job) error {
	slog.Info("starting kubernetes job", "job_name", job.GetJobName(), "template", job.JobTemplatePath())

	subDir := fmt.Sprintf("jobs/%s", job.JobTemplatePath())

	type resourceDef struct {
		FileSuffix string
		Process    func(rendered string) error
	}

	ctx := context.TODO()

	resources := []resourceDef{
		{
			FileSuffix: "_job.yaml",
			Process: func(rendered string) error {
				var jobObj batchv1.Job
				if err := yaml.Unmarshal([]byte(rendered), &jobObj); err != nil {
					return fmt.Errorf("error unmarshaling job YAML: %w", err)
				}
				if _, err := c.clientset.BatchV1().Jobs(c.namespace).Create(ctx, &jobObj, metav1.CreateOptions{}); err != nil {
					slog.Error("error creating job resource", "error", err)
					return fmt.Errorf("error creating job resource: %w", err)
				}
				return nil
			},
		},
		{
			FileSuffix: "_deployment.yaml",
			Process: func(rendered string) error {
				var deployment appsv1.Deployment
				if err := yaml.Unmarshal([]byte(rendered), &deployment); err != nil {
					return fmt.Errorf("error unmarshaling deployment YAML: %w", err)
				}
				if _, err := c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, &deployment, metav1.CreateOptions{}); err != nil {
					slog.Error("error creating deployment resource", "error", err)
					return fmt.Errorf("error creating deployment resource: %w", err)
				}
				return nil
			},
		},
		{
			FileSuffix: "_service.yaml",
			Process: func(rendered string) error {
				var service corev1.Service
				if err := yaml.Unmarshal([]byte(rendered), &service); err != nil {
					return fmt.Errorf("error unmarshaling service YAML: %w", err)
				}
				if _, err := c.clientset.CoreV1().Services(c.namespace).Create(ctx, &service, metav1.CreateOptions{}); err != nil {
					slog.Error("error creating service resource", "error", err)
					return fmt.Errorf("error creating service resource: %w", err)
				}
				return nil
			},
		},
		{
			FileSuffix: "_ingress.yaml",
			Process: func(rendered string) error {
				var ingress networkingv1.Ingress
				if err := yaml.Unmarshal([]byte(rendered), &ingress); err != nil {
					return fmt.Errorf("error unmarshaling ingress YAML: %w", err)
				}
				if _, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Create(ctx, &ingress, metav1.CreateOptions{}); err != nil {
					slog.Error("error creating ingress resource", "error", err)
					return fmt.Errorf("error creating ingress resource: %w", err)
				}
				return nil
			},
		},
		// TODO: add autoscaler process
	}

	processTemplate := func(fileSuffix string, process func(rendered string) error) error {
		templatePath := filepath.Join(subDir, job.JobTemplatePath()+fileSuffix)
		content, err := fs.ReadFile(jobTemplates, templatePath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				slog.Info("template file not found, skipping", "template", templatePath)
				return nil
			}
			slog.Error("error reading template file", "template", templatePath, "error", err)
			return fmt.Errorf("error reading template file %s: %w", templatePath, err)
		}

		tmpl, err := template.New(job.JobTemplatePath() + fileSuffix).Parse(string(content))
		if err != nil {
			slog.Error("error parsing template", "template", templatePath, "error", err)
			return fmt.Errorf("error parsing template %s: %w", templatePath, err)
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, job); err != nil {
			slog.Error("error rendering template", "template", templatePath, "error", err)
			return fmt.Errorf("error rendering template %s: %w", templatePath, err)
		}
		rendered := buf.String()

		if err := process(rendered); err != nil {
			return fmt.Errorf("error submitting template %s: %w", templatePath, err)
		}
		slog.Info("resource created successfully", "template", templatePath)
		return nil
	}

	for _, res := range resources {
		if err := processTemplate(res.FileSuffix, res.Process); err != nil {
			return err
		}
	}

	slog.Info("all resources for job started successfully", "job_name", job.GetJobName())
	return nil
}

func (c *KubernetesClient) StopJob(jobName string) error {
	slog.Info("stopping kubernetes job resources", "job_name", jobName)
	var errs []string
	ctx := context.TODO()

	if err := c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		slog.Error("error stopping deployment", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Sprintf("deployment: %v", err))
	}

	if err := c.clientset.CoreV1().Services(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		slog.Error("error stopping service", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Sprintf("service: %v", err))
	}

	if err := c.clientset.NetworkingV1().Ingresses(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		slog.Error("error stopping ingress", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Sprintf("ingress: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("error stopping job resources for %s: %s", jobName, strings.Join(errs, "; "))
	}
	slog.Info("kubernetes job resources stopped successfully", "job_name", jobName)
	return nil
}

func (c *KubernetesClient) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	ctx := context.TODO()

	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		var status string
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 {
			if deployment.Status.AvailableReplicas > 0 {
				status = "running"
			} else {
				status = "pending"
			}
		} else {
			status = "dead"
		}
		info := orchestrator.JobInfo{
			Name:   deployment.Name,
			Status: status,
		}
		slog.Info("job info retrieved from deployment", "job_name", jobName, "status", status)
		return info, nil
	}

	if !apierrors.IsNotFound(err) {
		slog.Error("error getting deployment", "job_name", jobName, "error", err)
		return orchestrator.JobInfo{}, fmt.Errorf("error getting deployment %s: %w", jobName, err)
	}

	job, err := c.clientset.BatchV1().Jobs(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		slog.Error("error getting job", "job_name", jobName, "error", err)
		return orchestrator.JobInfo{}, fmt.Errorf("error getting job %s: %w", jobName, err)
	}

	var status string
	if job.Status.Active > 0 {
		status = "running"
	} else if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
		status = "dead"
	} else {
		status = "pending"
	}
	info := orchestrator.JobInfo{
		Name:   job.Name,
		Status: status,
	}
	slog.Info("kubernetes job info retrieved", "job_name", jobName, "status", status)
	return info, nil
}

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
