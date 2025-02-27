package kubernetes

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// For Job manifests.
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"thirdai_platform/model_bazaar/orchestrator"
)

//go:embed jobs/*/*
var jobTemplates embed.FS

func getNamespace() (string, error) {
	slog.Info("reading namespace from file", "path", "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		slog.Info("failed to read namespace file", "error", err)
		return "", err
	}
	namespace := strings.TrimSpace(string(data))
	slog.Info("namespace retrieved", "namespace", namespace)
	return namespace, nil
}

type KubernetesClient struct {
	namespace       string
	clientset       *kubernetes.Clientset
	ingressHostname string
}

func NewKubernetesClient(ingressHostname string) orchestrator.Client {
	var config *rest.Config
	var err error
	slog.Info("initializing NewKubernetesClient", "ingressHostname", ingressHostname)

	platform := os.Getenv("PLATFORM")

	slog.Info("PLATFORM is not set to 'local', using in-cluster Kubernetes config...")

	config, err = rest.InClusterConfig()
	if err != nil {
		log.Panicf("error getting in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("error creating kubernetes clientset", "error", err)
		panic(fmt.Sprintf("error creating kubernetes clientset: %v", err))
	}

	var namespace string
	if platform == "local" {
		namespace = "default"
	} else {
		namespace, err = getNamespace()
		if err != nil {
			log.Panicf("error retrieving kubernetes namespace: %v", err)
		}
	}

	slog.Info("creating kubernetes client", "namespace", namespace, "ingressHostname", ingressHostname)
	return &KubernetesClient{
		namespace:       namespace,
		clientset:       clientset,
		ingressHostname: ingressHostname,
	}
}

func (c *KubernetesClient) StartJob(job orchestrator.Job) error {
	slog.Info("starting kubernetes job", "job_name", job.GetJobName(), "template", job.JobTemplatePath(), "namespace", c.namespace)
	subDir := fmt.Sprintf("jobs/%s", job.JobTemplatePath())

	ctx := context.Background()

	var renderedJobYAML strings.Builder
	var err error

	for _, res := range resources {
		slog.Info("processing resource type", "fileSuffix", res.FileSuffix, "job_name", job.GetJobName())
		renderedJobYAML, err = c.processTemplate(res.FileSuffix, subDir, job, ctx, renderedJobYAML)
		if err != nil {
			slog.Error("error processing template", "fileSuffix", res.FileSuffix, "job_name", job.GetJobName(), "error", err)
			return err
		}
	}
	shareDir := os.Getenv("SHARE_DIR")

	outputFile := filepath.Join(shareDir, "jobs", fmt.Sprintf("%s.yaml", job.GetJobName()))
	if err := os.WriteFile(outputFile, []byte(renderedJobYAML.String()), 0644); err != nil {
		return fmt.Errorf("error writing job YAML file: %w", err)
	}

	slog.Info("all resources for job started successfully", "job_name", job.GetJobName(), "namespace", c.namespace)
	return nil
}

func (c *KubernetesClient) StopJob(jobName string) error {
	slog.Info("stopping kubernetes job resources", "job_name", jobName, "namespace", c.namespace)
	var errs []error
	ctx := context.Background()

	// Delete deployment (assumed to use the jobName)
	slog.Info("attempting to delete deployment", "deployment_name", jobName, "namespace", c.namespace)
	if err := c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			slog.Info("deployment not found, skipping deletion", "deployment_name", jobName)
		} else {
			slog.Error("error stopping deployment", "deployment_name", jobName, "error", err)
			errs = append(errs, fmt.Errorf("deployment: %w", err))
		}
	} else {
		slog.Info("deployment deleted successfully", "deployment_name", jobName)
	}

	// Delete service (assumed to use the jobName)
	slog.Info("attempting to delete service", "service_name", jobName, "namespace", c.namespace)
	if err := c.clientset.CoreV1().Services(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			slog.Info("service not found, skipping deletion", "service_name", jobName)
		} else {
			slog.Error("error stopping service", "service_name", jobName, "error", err)
			errs = append(errs, fmt.Errorf("service: %w", err))
		}
	} else {
		slog.Info("service deleted successfully", "service_name", jobName)
	}

	// Delete external ingress (assumed to use the jobName)
	slog.Info("attempting to delete ingress", "ingress_name", jobName, "namespace", c.namespace)
	if err := c.clientset.NetworkingV1().Ingresses(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			slog.Info("ingress not found, skipping deletion", "ingress_name", jobName)
		} else {
			slog.Error("error stopping ingress", "ingress_name", jobName, "error", err)
			errs = append(errs, fmt.Errorf("ingress: %w", err))
		}
	} else {
		slog.Info("ingress deleted successfully", "ingress_name", jobName)
	}

	// Delete internal ingress with suffix "-internal"
	internalIngressName := jobName + "-internal"
	slog.Info("attempting to delete internal ingress", "ingress_name", internalIngressName, "namespace", c.namespace)
	if err := c.clientset.NetworkingV1().Ingresses(c.namespace).Delete(ctx, internalIngressName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			slog.Info("internal ingress not found, skipping deletion", "ingress_name", internalIngressName)
		} else {
			slog.Error("error stopping internal ingress", "ingress_name", internalIngressName, "error", err)
			errs = append(errs, fmt.Errorf("internal ingress: %w", err))
		}
	} else {
		slog.Info("internal ingress deleted successfully", "ingress_name", internalIngressName)
	}
	shareDir := os.Getenv("SHARE_DIR")
	outputFile := filepath.Join(shareDir, "jobs", fmt.Sprintf("%s.yaml", jobName))
	if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
		slog.Error("error deleting job YAML file", "job_name", jobName, "error", err)
		errs = append(errs, fmt.Errorf("job yaml deletion: %v", err))
	}

	if len(errs) > 0 {
		slog.Info("one or more errors occurred while stopping job resources", "job_name", jobName, "errors", errs)
		return fmt.Errorf("error stopping job resources for %s: %w", jobName, errors.Join(errs...))
	}
	slog.Info("kubernetes job resources stopped successfully", "job_name", jobName, "namespace", c.namespace)
	return nil
}

func (c *KubernetesClient) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	slog.Info("retrieving job info", "job_name", jobName, "namespace", c.namespace)
	ctx := context.Background()

	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		slog.Info("deployment found for job", "job_name", jobName)
		var status orchestrator.JobStatus
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 {
			if deployment.Status.AvailableReplicas > 0 {
				status = orchestrator.StatusRunning
			} else {
				status = orchestrator.StatusPending
			}
		} else {
			status = orchestrator.StatusDead
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

	slog.Info("deployment not found, trying batch job for", "job_name", jobName)
	job, err := c.clientset.BatchV1().Jobs(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			slog.Info("kubernetes job not found")
			return orchestrator.JobInfo{}, orchestrator.ErrJobNotFound
		}
		slog.Error("error getting job", "job_name", jobName, "error", err)
		return orchestrator.JobInfo{}, fmt.Errorf("error getting job %s: %w", jobName, err)
	}

	var status orchestrator.JobStatus
	if job.Status.Active > 0 {
		status = orchestrator.StatusRunning
	} else if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
		status = orchestrator.StatusDead
	} else {
		status = orchestrator.StatusPending
	}
	info := orchestrator.JobInfo{
		Name:   job.Name,
		Status: status,
	}
	slog.Info("kubernetes job info retrieved", "job_name", jobName, "status", status)
	return info, nil
}

func (c *KubernetesClient) JobLogs(jobName string) ([]orchestrator.JobLog, error) {
	slog.Info("retrieving job logs", "job_name", jobName, "namespace", c.namespace)
	ctx := context.Background()
	podList, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", jobName),
	})
	if err != nil {
		slog.Error("error listing pods for job", "job_name", jobName, "error", err)
		return nil, fmt.Errorf("error listing pods for job %s: %w", jobName, err)
	}
	slog.Info("pods listed for job", "job_name", jobName, "podCount", len(podList.Items))

	var logs []orchestrator.JobLog
	for _, pod := range podList.Items {
		slog.Info("retrieving logs for pod", "pod", pod.Name)
		podLog, err := c.getPodLogs(pod.Name)
		if err != nil {
			slog.Error("error getting logs for pod", "pod", pod.Name, "error", err)
			return nil, err
		}
		logs = append(logs, orchestrator.JobLog{
			Stdout: podLog,
			Stderr: "",
		})
		slog.Info("logs retrieved for pod", "pod", pod.Name)
	}
	slog.Info("job logs retrieved", "job_name", jobName)
	return logs, nil
}

func (c *KubernetesClient) getPodLogs(podName string) (string, error) {
	slog.Info("opening log stream for pod", "podName", podName, "namespace", c.namespace)
	ctx := context.Background()
	podLogOpts := corev1.PodLogOptions{}
	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, &podLogOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		slog.Error("error opening log stream", "podName", podName, "error", err)
		return "", fmt.Errorf("error opening log stream for pod %s: %w", podName, err)
	}
	defer stream.Close()
	slog.Info("log stream opened", "podName", podName)

	var builder strings.Builder
	_, err = io.Copy(&builder, stream)
	if err != nil {
		slog.Error("error reading log stream", "podName", podName, "error", err)
		return "", fmt.Errorf("error reading log stream for pod %s: %w", podName, err)
	}
	slog.Info("log stream read completed", "podName", podName)
	return builder.String(), nil
}

func (c *KubernetesClient) ListServices() ([]orchestrator.ServiceInfo, error) {
	slog.Info("listing services", "namespace", c.namespace)
	ctx := context.Background()
	svcList, err := c.clientset.CoreV1().Services(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("error listing services", "namespace", c.namespace, "error", err)
		return nil, fmt.Errorf("error listing services: %w", err)
	}

	var services []orchestrator.ServiceInfo
	for _, svc := range svcList.Items {
		slog.Info("processing service", "service_name", svc.Name, "namespace", c.namespace)
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
					slog.Info("service allocation added", "service", svc.Name, "address", addr.IP, "port", port.Port)
				}
			}
		}

		services = append(services, orchestrator.ServiceInfo{
			Name:        svc.Name,
			Allocations: allocations,
		})
	}
	slog.Info("services listed", "count", len(services), "namespace", c.namespace)
	return services, nil
}

func (c *KubernetesClient) TotalCpuUsage() (int, error) {
	slog.Info("calculating total CPU usage", "namespace", c.namespace)
	ctx := context.Background()
	podList, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("error listing pods", "namespace", c.namespace, "error", err)
		return 0, fmt.Errorf("error listing pods: %w", err)
	}

	totalMillicores := 0
	for _, pod := range podList.Items {
		slog.Info("processing pod for CPU usage", "pod", pod.Name, "phase", pod.Status.Phase)
		if pod.Status.Phase != corev1.PodRunning {
			slog.Info("skipping pod not in Running phase", "pod", pod.Name, "phase", pod.Status.Phase)
			continue
		}
		for _, container := range pod.Spec.Containers {
			if cpu, ok := container.Resources.Requests["cpu"]; ok {
				slog.Info("found CPU request", "pod", pod.Name, "container", container.Name, "cpuRequest", cpu.String())
				m, err := parseCPUQuantity(cpu.String())
				if err != nil {
					slog.Error("error parsing cpu quantity", "pod", pod.Name, "quantity", cpu.String(), "error", err)
					continue
				}
				totalMillicores += m
			}
		}
	}
	slog.Info("total CPU usage calculated", "totalMillicores", totalMillicores, "namespace", c.namespace)
	return totalMillicores, nil
}

func parseCPUQuantity(q string) (int, error) {
	slog.Info("parsing CPU quantity", "quantity", q)
	if strings.HasSuffix(q, "m") {
		trimmed := strings.TrimSuffix(q, "m")
		m, err := strconv.Atoi(trimmed)
		if err != nil {
			slog.Error("error converting milli value", "quantity", q, "error", err)
			return 0, err
		}
		slog.Info("parsed milli CPU quantity", "quantity", q, "millicores", m)
		return m, nil
	}
	f, err := strconv.ParseFloat(q, 64)
	if err != nil {
		slog.Error("error converting CPU quantity to float", "quantity", q, "error", err)
		return 0, err
	}
	millicores := int(f * 1000)
	slog.Info("parsed CPU quantity", "quantity", q, "millicores", millicores)
	return millicores, nil
}

func (c *KubernetesClient) IngressHostname() string {
	slog.Info("returning ingress hostname", "ingressHostname", c.ingressHostname)
	return c.ingressHostname
}
