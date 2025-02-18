package kubernetes

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	appsv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1" // For Job manifests.
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	slog.Info("initializing NewKubernetesClient", "ingressHostname", ingressHostname)

	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("error getting in-cluster config", "error", err)
		panic(fmt.Sprintf("error getting in-cluster config: %v", err))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("error creating kubernetes clientset", "error", err)
		panic(fmt.Sprintf("error creating kubernetes clientset: %v", err))
	}

	namespace, err := getNamespace()
	if err != nil {
		slog.Error("error retrieving kubernetes namespace", "error", err)
		panic(fmt.Sprintf("error retrieving kubernetes namespace: %v", err))
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

	type resourceDef struct {
		FileSuffix string
		Process    func(rendered string) error
	}

	ctx := context.TODO()

	// For each resource type, we define a Process function that:
	// 1. Unmarshals the YAML into the appropriate object.
	// 2. Checks whether the resource already exists.
	// 3a. If it does not exist, creates it.
	// 3b. If it exists and supports an update, we set the ResourceVersion
	//     and call Update. If it is better replaced, we delete the old
	//     resource and then create the new one.
	resources := []resourceDef{
		{
			FileSuffix: "_job.yaml",
			Process: func(doc string) error {
				slog.Info("processing job YAML document", "namespace", c.namespace)
				var jobObj batchv1.Job
				if err := yaml.Unmarshal([]byte(doc), &jobObj); err != nil {
					return fmt.Errorf("error unmarshaling job YAML: %w", err)
				}
				slog.Info("job YAML unmarshaled", "job_name", jobObj.Name)

				slog.Info("checking if job resource exists", "job_name", jobObj.Name, "namespace", c.namespace)
				_, err := c.clientset.BatchV1().Jobs(c.namespace).Get(ctx, jobObj.Name, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						slog.Info("job resource not found, creating new job", "job_name", jobObj.Name)
						if _, err := c.clientset.BatchV1().Jobs(c.namespace).Create(ctx, &jobObj, metav1.CreateOptions{}); err != nil {
							slog.Error("error creating job resource", "job_name", jobObj.Name, "error", err)
							return fmt.Errorf("error creating job resource: %w", err)
						}
						slog.Info("job resource created successfully", "job_name", jobObj.Name)
						return nil
					}
					return fmt.Errorf("error checking for existing job: %w", err)
				}

				slog.Info("job resource exists, deleting it", "job_name", jobObj.Name)
				if err := c.clientset.BatchV1().Jobs(c.namespace).Delete(ctx, jobObj.Name, metav1.DeleteOptions{}); err != nil {
					slog.Error("error deleting existing job resource", "job_name", jobObj.Name, "error", err)
					return fmt.Errorf("error deleting existing job resource: %w", err)
				}
				slog.Info("re-creating job resource after deletion", "job_name", jobObj.Name)
				if _, err := c.clientset.BatchV1().Jobs(c.namespace).Create(ctx, &jobObj, metav1.CreateOptions{}); err != nil {
					slog.Error("error re-creating job resource", "job_name", jobObj.Name, "error", err)
					return fmt.Errorf("error re-creating job resource: %w", err)
				}
				slog.Info("job resource re-created successfully", "job_name", jobObj.Name)
				return nil
			},
		},
		{
			FileSuffix: "_deployment.yaml",
			Process: func(doc string) error {
				slog.Info("processing deployment YAML document", "namespace", c.namespace)
				var deployment appsv1.Deployment
				if err := yaml.Unmarshal([]byte(doc), &deployment); err != nil {
					return fmt.Errorf("error unmarshaling deployment YAML: %w", err)
				}
				slog.Info("deployment YAML unmarshaled", "deployment_name", deployment.Name)

				slog.Info("checking if deployment resource exists", "deployment_name", deployment.Name, "namespace", c.namespace)
				existing, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						slog.Info("deployment resource not found, creating new deployment", "deployment_name", deployment.Name)
						if _, err := c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, &deployment, metav1.CreateOptions{}); err != nil {
							slog.Error("error creating deployment resource", "deployment_name", deployment.Name, "error", err)
							return fmt.Errorf("error creating deployment resource: %w", err)
						}
						slog.Info("deployment resource created successfully", "deployment_name", deployment.Name)
						return nil
					}
					return fmt.Errorf("error checking for existing deployment: %w", err)
				}

				slog.Info("deployment resource exists, updating deployment", "deployment_name", deployment.Name)
				deployment.ResourceVersion = existing.ResourceVersion
				if _, err := c.clientset.AppsV1().Deployments(c.namespace).Update(ctx, &deployment, metav1.UpdateOptions{}); err != nil {
					slog.Error("error updating deployment resource", "deployment_name", deployment.Name, "error", err)
					return fmt.Errorf("error updating deployment resource: %w", err)
				}
				slog.Info("deployment resource updated successfully", "deployment_name", deployment.Name)
				return nil
			},
		},
		{
			FileSuffix: "_service.yaml",
			Process: func(doc string) error {
				slog.Info("processing service YAML document", "namespace", c.namespace)
				var service corev1.Service
				if err := yaml.Unmarshal([]byte(doc), &service); err != nil {
					return fmt.Errorf("error unmarshaling service YAML: %w", err)
				}
				slog.Info("service YAML unmarshaled", "service_name", service.Name)

				slog.Info("checking if service resource exists", "service_name", service.Name, "namespace", c.namespace)
				existing, err := c.clientset.CoreV1().Services(c.namespace).Get(ctx, service.Name, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						slog.Info("service resource not found, creating new service", "service_name", service.Name)
						if _, err := c.clientset.CoreV1().Services(c.namespace).Create(ctx, &service, metav1.CreateOptions{}); err != nil {
							slog.Error("error creating service resource", "service_name", service.Name, "error", err)
							return fmt.Errorf("error creating service resource: %w", err)
						}
						slog.Info("service resource created successfully", "service_name", service.Name)
						return nil
					}
					return fmt.Errorf("error checking for existing service: %w", err)
				}

				slog.Info("service resource exists, updating service", "service_name", service.Name)
				service.ResourceVersion = existing.ResourceVersion
				if _, err := c.clientset.CoreV1().Services(c.namespace).Update(ctx, &service, metav1.UpdateOptions{}); err != nil {
					slog.Error("error updating service resource", "service_name", service.Name, "error", err)
					return fmt.Errorf("error updating service resource: %w", err)
				}
				slog.Info("service resource updated successfully", "service_name", service.Name)
				return nil
			},
		},
		{
			FileSuffix: "_ingress.yaml",
			Process: func(doc string) error {
				slog.Info("processing ingress YAML document", "namespace", c.namespace)
				var ingress networkingv1.Ingress
				if err := yaml.Unmarshal([]byte(doc), &ingress); err != nil {
					return fmt.Errorf("error unmarshaling ingress YAML: %w", err)
				}
				slog.Info("ingress YAML unmarshaled", "ingress_name", ingress.Name)

				slog.Info("checking if ingress resource exists", "ingress_name", ingress.Name, "namespace", c.namespace)
				existing, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Get(ctx, ingress.Name, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						slog.Info("ingress resource not found, creating new ingress", "ingress_name", ingress.Name)
						if _, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Create(ctx, &ingress, metav1.CreateOptions{}); err != nil {
							slog.Error("error creating ingress resource", "ingress_name", ingress.Name, "error", err)
							return fmt.Errorf("error creating ingress resource: %w", err)
						}
						slog.Info("ingress resource created successfully", "ingress_name", ingress.Name)
						return nil
					}
					return fmt.Errorf("error checking for existing ingress: %w", err)
				}

				slog.Info("ingress resource exists, updating ingress", "ingress_name", ingress.Name)
				ingress.ResourceVersion = existing.ResourceVersion
				if _, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Update(ctx, &ingress, metav1.UpdateOptions{}); err != nil {
					slog.Error("error updating ingress resource", "ingress_name", ingress.Name, "error", err)
					return fmt.Errorf("error updating ingress resource: %w", err)
				}
				slog.Info("ingress resource updated successfully", "ingress_name", ingress.Name)
				return nil
			},
		},
		{
			FileSuffix: "_hpa.yaml",
			Process: func(doc string) error {
				slog.Info("Processing HPA YAML document", "namespace", c.namespace)
				var hpa v2.HorizontalPodAutoscaler
				if err := yaml.Unmarshal([]byte(doc), &hpa); err != nil {
					return fmt.Errorf("error unmarshaling HPA YAML: %w", err)
				}
				slog.Info("HPA YAML unmarshaled", "hpa_name", hpa.Name)

				existing, err := c.clientset.AutoscalingV2().HorizontalPodAutoscalers(c.namespace).Get(ctx, hpa.Name, metav1.GetOptions{})
				if err != nil {
					if apierrors.IsNotFound(err) {
						slog.Info("HPA resource not found, creating new HPA", "hpa_name", hpa.Name)
						if _, err := c.clientset.AutoscalingV2().HorizontalPodAutoscalers(c.namespace).Create(ctx, &hpa, metav1.CreateOptions{}); err != nil {
							slog.Error("Error creating HPA resource", "hpa_name", hpa.Name, "error", err)
							return fmt.Errorf("error creating HPA resource: %w", err)
						}
						slog.Info("HPA resource created successfully", "hpa_name", hpa.Name)
						return nil
					}
					return fmt.Errorf("error checking for existing HPA: %w", err)
				}

				// HPA exists, so update it.
				slog.Info("HPA resource exists, updating HPA", "hpa_name", hpa.Name)
				hpa.ResourceVersion = existing.ResourceVersion
				if _, err := c.clientset.AutoscalingV2().HorizontalPodAutoscalers(c.namespace).Update(ctx, &hpa, metav1.UpdateOptions{}); err != nil {
					slog.Error("Error updating HPA resource", "hpa_name", hpa.Name, "error", err)
					return fmt.Errorf("error updating HPA resource: %w", err)
				}
				slog.Info("HPA resource updated successfully", "hpa_name", hpa.Name)
				return nil
			},
		},
	}

	// processTemplate reads the file, renders the template with the job’s data,
	// and then calls the resource-specific process function.
	processTemplate := func(fileSuffix string, process func(rendered string) error) error {
		templatePath := filepath.Join(subDir, job.JobTemplatePath()+fileSuffix)
		slog.Info("processing template", "templatePath", templatePath, "fileSuffix", fileSuffix, "job_name", job.GetJobName(), "namespace", c.namespace)
		content, err := fs.ReadFile(jobTemplates, templatePath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				slog.Info("template file not found, skipping", "template", templatePath)
				return nil
			}
			slog.Error("error reading template file", "template", templatePath, "error", err)
			return fmt.Errorf("error reading template file %s: %w", templatePath, err)
		}

		tmpl, err := template.New(job.JobTemplatePath() + fileSuffix).
			Funcs(template.FuncMap{
				"namespace": func() string {
					return c.namespace
				},
			}).
			Parse(string(content))
		if err != nil {
			slog.Error("error parsing template", "template", templatePath, "error", err)
			return fmt.Errorf("error parsing template %s: %w", templatePath, err)
		}
		slog.Info("template parsed successfully", "template", templatePath)

		var buf strings.Builder
		if err := tmpl.Execute(&buf, job); err != nil {
			slog.Error("error rendering template", "template", templatePath, "error", err)
			return fmt.Errorf("error rendering template %s: %w", templatePath, err)
		}
		rendered := buf.String()
		slog.Info("template rendered", "template", templatePath)

		// Process multiple docs in a single YAML file
		decoder := yaml.NewDecoder(strings.NewReader(rendered))
		for {
			var doc interface{}
			if err := decoder.Decode(&doc); err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("error decoding YAML document: %w", err)
			}

			if doc == nil {
				continue
			}

			slog.Info("processing individual YAML document", "template", templatePath)

			docBytes, err := yaml.Marshal(doc)
			if err != nil {
				return fmt.Errorf("error marshaling YAML document: %w", err)
			}

			if err := process(string(docBytes)); err != nil {
				return fmt.Errorf("error submitting template %s: %w", templatePath, err)
			}
		}
		slog.Info("resources created/updated successfully", "template", templatePath)
		return nil
	}

	// Process all defined resource templates.
	for _, res := range resources {
		slog.Info("processing resource type", "fileSuffix", res.FileSuffix, "job_name", job.GetJobName())
		if err := processTemplate(res.FileSuffix, res.Process); err != nil {
			slog.Error("error processing template", "fileSuffix", res.FileSuffix, "job_name", job.GetJobName(), "error", err)
			return err
		}
	}

	slog.Info("all resources for job started successfully", "job_name", job.GetJobName(), "namespace", c.namespace)
	return nil
}

func (c *KubernetesClient) StopJob(jobName string) error {
	slog.Info("stopping kubernetes job resources", "job_name", jobName, "namespace", c.namespace)
	var errs []error
	ctx := context.TODO()

	// Delete deployment with suffix "-deployment"
	deploymentName := jobName + "-deployment"
	slog.Info("attempting to delete deployment", "deployment_name", deploymentName, "namespace", c.namespace)
	if err := c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			slog.Info("deployment not found, skipping deletion", "deployment_name", deploymentName)
		} else {
			slog.Error("error stopping deployment", "deployment_name", deploymentName, "error", err)
			errs = append(errs, fmt.Errorf("deployment: %w", err))
		}
	} else {
		slog.Info("deployment deleted successfully", "deployment_name", deploymentName)
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

	if len(errs) > 0 {
		slog.Info("one or more errors occurred while stopping job resources", "job_name", jobName, "errors", errs)
		return fmt.Errorf("error stopping job resources for %s: %w", jobName, errors.Join(errs...))
	}
	slog.Info("kubernetes job resources stopped successfully", "job_name", jobName, "namespace", c.namespace)
	return nil
}

func (c *KubernetesClient) JobInfo(jobName string) (orchestrator.JobInfo, error) {
	slog.Info("retrieving job info", "job_name", jobName, "namespace", c.namespace)
	ctx := context.TODO()

	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err == nil {
		slog.Info("deployment found for job", "job_name", jobName)
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
	slog.Info("retrieving job logs", "job_name", jobName, "namespace", c.namespace)
	ctx := context.TODO()
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
	ctx := context.TODO()
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
	ctx := context.TODO()
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
	ctx := context.TODO()
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
