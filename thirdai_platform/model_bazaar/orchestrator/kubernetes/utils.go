package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"text/template"
	"thirdai_platform/model_bazaar/orchestrator"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

type resourceDef struct {
	FileSuffix   string
	ResourceType string
}

var resources = []resourceDef{
	{
		FileSuffix:   "_job.yaml",
		ResourceType: "Job",
	},
	{
		FileSuffix:   "_deployment.yaml",
		ResourceType: "Deployment",
	},
	{
		FileSuffix:   "_service.yaml",
		ResourceType: "Service",
	},
	{
		FileSuffix:   "_ingress.yaml",
		ResourceType: "Ingress",
	},
	{
		FileSuffix:   "_hpa.yaml",
		ResourceType: "HorizontalPodAutoscaler",
	},
	{
		FileSuffix:   "_rba.yaml",
		ResourceType: "RoleBinding",
	},
}

func (c *KubernetesClient) processJob(doc string, ctx context.Context) error {
	slog.Info("processing job YAML document", "namespace", c.namespace)
	var jobObj batchv1.Job
	if err := k8syaml.Unmarshal([]byte(doc), &jobObj); err != nil {
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
}

func (c *KubernetesClient) processDeployment(doc string, ctx context.Context) error {
	slog.Info("processing deployment YAML document", "namespace", c.namespace)
	var deployment appsv1.Deployment
	if err := k8syaml.Unmarshal([]byte(doc), &deployment); err != nil {
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
}

func (c *KubernetesClient) processService(doc string, ctx context.Context) error {
	slog.Info("processing service YAML document", "namespace", c.namespace)
	var service corev1.Service
	if err := k8syaml.Unmarshal([]byte(doc), &service); err != nil {
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
}

func (c *KubernetesClient) processIngress(doc string, ctx context.Context) error {
	slog.Info("processing ingress YAML document", "namespace", c.namespace)
	var ingress networkingv1.Ingress
	if err := k8syaml.Unmarshal([]byte(doc), &ingress); err != nil {
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
}

func (c *KubernetesClient) processHPA(doc string, ctx context.Context) error {
	slog.Info("Processing HPA YAML document", "namespace", c.namespace)
	var hpa v2.HorizontalPodAutoscaler
	if err := k8syaml.Unmarshal([]byte(doc), &hpa); err != nil {
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
}

func (c *KubernetesClient) processByFileSuffix(fileSuffix string, doc string, ctx context.Context) error {
	switch fileSuffix {
	case "_job.yaml":
		return c.processJob(doc, ctx)
	case "_deployment.yaml":
		return c.processDeployment(doc, ctx)
	case "_service.yaml":
		return c.processService(doc, ctx)
	case "_ingress.yaml":
		return c.processIngress(doc, ctx)
	case "_hpa.yaml":
		return c.processHPA(doc, ctx)
	default:
		return fmt.Errorf("error processing template due to unknown file suffix: %s", fileSuffix)
	}
}

func (c *KubernetesClient) processTemplate(fileSuffix, subDir string, job orchestrator.Job, ctx context.Context) error {
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

		if err := c.processByFileSuffix(fileSuffix, string(docBytes), ctx); err != nil {
			return fmt.Errorf("error submitting template %s: %w", templatePath, err)
		}
	}
	slog.Info("resources created/updated successfully", "template", templatePath)
	return nil
}
