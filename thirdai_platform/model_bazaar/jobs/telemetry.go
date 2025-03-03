package jobs

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"thirdai_platform/model_bazaar/orchestrator"
	"thirdai_platform/model_bazaar/storage"

	"gopkg.in/yaml.v3"
)

// This will load the given templates into the embed FS so that they are bunddled
// into the compiled binary.

//go:embed grafana_dashboards/*
var grafanaDashboards embed.FS

func copyGrafanaDashboards(storage storage.Storage, orchestratorName string) error {
	dashboardDest := "cluster-monitoring/grafana/dashboards"

	exists, err := storage.Exists(dashboardDest)
	if err != nil {
		return fmt.Errorf("error checking if grafana dashboards exists: %w", err)
	}

	if exists {
		err := storage.Delete(dashboardDest)
		if err != nil {
			return fmt.Errorf("error deleting existing grafana dashboards directory: %w", err)
		}
	}

	return fs.WalkDir(grafanaDashboards, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking grafana dashboards: %w", err)
		}

		// Skip directories for the other orchestrator
		if orchestratorName != "kubernetes" && strings.HasSuffix(path, "kubernetes") {
			return fs.SkipDir
		} else if orchestratorName != "nomad" && strings.HasSuffix(path, "nomad") {
			return fs.SkipDir
		}

		relPath, err := filepath.Rel("grafana_dashboards", path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}
		destPath := filepath.Join(dashboardDest, relPath)
		if d.IsDir() {
			// directories would get created by storage.writeData
			return nil
		}
		content, err := fs.ReadFile(grafanaDashboards, path)
		if err != nil {
			return fmt.Errorf("error reading file %s from embedded filesystem: %w", path, err)
		}
		err = storage.Write(destPath, bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("error writing file %s to shared storage: %w", destPath, err)
		}
		return nil
	})
}

type TelemetryJobArgs struct {
	IsLocal             bool
	ModelBazaarEndpoint string
	Docker              orchestrator.DockerEnv
	GrafanaDbUrl        string
	AdminUsername       string
	AdminEmail          string
	AdminPassword       string
}

func StartTelemetryJob(orchestratorClient orchestrator.Client, storage storage.Storage, args TelemetryJobArgs) error {
	slog.Info("starting telemetry job")

	// create prometheus config file
	err := createPromFile(orchestratorClient.GetName(), storage, args.ModelBazaarEndpoint, args.IsLocal)
	if err != nil {
		return fmt.Errorf("error creating promfile: %w", err)
	}

	// copy grafana dashboards to appropriate directory
	err = copyGrafanaDashboards(storage, orchestratorClient.GetName())
	if err != nil {
		slog.Error("error initializing grafana dashboards", "error", err)
		return fmt.Errorf("error initializing grafana dashboards: %w", err)
	}

	//create grafana provisioning configs
	err = createGrafanaProvisionings(storage, args.IsLocal, orchestratorClient.GetName(), args.ModelBazaarEndpoint)
	if err != nil {
		return fmt.Errorf("error creating grafana provisioning: %w", err)
	}

	// create vector config file
	err = createVectorConfig(storage, args.ModelBazaarEndpoint)
	if err != nil {
		return fmt.Errorf("error creating vector config file: %w", err)
	}

	job := orchestrator.TelemetryJob{
		IsLocal:              args.IsLocal,
		ClusterMonitoringDir: filepath.Join(storage.Location(), "cluster-monitoring"),
		AdminUsername:        args.AdminUsername,
		AdminEmail:           args.AdminEmail,
		AdminPassword:        args.AdminPassword,
		GrafanaDbUrl:         args.GrafanaDbUrl,
		Docker:               args.Docker,
		IngressHostname:      orchestratorClient.IngressHostname(),
	}

	if args.IsLocal {
		// When running in production with docker we don't restart here because multiple
		// model bazaar jobs could be used. When an installation is updated the docker
		// version will be updated which will cause nomad to detect a change in the hcl
		// file and thus restart the job when StartJob is invoked later. If multiple
		// model bazaar jobs call StartJob with the same version, nomad will ignore
		// subsequent calls.
		err := orchestrator.StopJobIfExists(orchestratorClient, job.GetJobName())
		if err != nil {
			slog.Error("error stopping existing telemetry job", "error", err)
			return fmt.Errorf("error stopping existing telemetry job: %w", err)
		}
	}

	err = orchestratorClient.StartJob(job)
	if err != nil {
		slog.Error("error starting telemetry job", "error", err)
		return fmt.Errorf("error starting telemetry job: %w", err)
	}

	slog.Info("telemetry job started successfully")
	return nil
}

type targetList struct {
	Targets []string
	Labels  map[string]string
}

func createPromFile(orchestratorName string, storage storage.Storage, modelBazaarEndpoint string, isLocal bool) error {
	serverNodeFile := filepath.Join("cluster-monitoring", "nomad_nodes", "server.yaml")

	if isLocal && orchestratorName == "nomad" {
		data, err := yaml.Marshal(
			[]targetList{{
				Targets: []string{"host.docker.internal:4646"},
				Labels:  map[string]string{"nomad_node": "server"},
			}},
		)
		if err != nil {
			return fmt.Errorf("error creating local server.yaml file: %w", err)
		}

		err = storage.Write(serverNodeFile, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("error writing local server.yaml file: %w", err)
		}
	}

	promfile, err := yaml.Marshal(prometheusConfig(orchestratorName, modelBazaarEndpoint, isLocal))
	if err != nil {
		return fmt.Errorf("error creating promfile: %w", err)
	}

	err = storage.Write(
		filepath.Join("cluster-monitoring", "node_discovery", "prometheus.yaml"),
		bytes.NewReader(promfile),
	)
	if err != nil {
		return fmt.Errorf("error writing promfile: %w", err)
	}

	return nil
}

func getDeploymentTargetsEndpoint(modelBazaarEndpoint string, isLocal bool, orchestratorName string) string {
	if isLocal && orchestratorName == "nomad" {
		return "http://host.docker.internal:80/api/v2/telemetry/deployment-services"
	} else {
		return strings.TrimSuffix(modelBazaarEndpoint, "/") + "/api/v2/telemetry/deployment-services"
	}
}

func prometheusConfig(orchestratorName string, modelBazaarEndpoint string, isLocal bool) map[string]interface{} {
	deploymentTargetsEndpoint := getDeploymentTargetsEndpoint(modelBazaarEndpoint, isLocal, orchestratorName)

	var orchestratorEntry []map[string]interface{}
	if orchestratorName == "nomad" {
		orchestratorEntry = []map[string]interface{}{
			{
				"job_name":     "nomad-agent",
				"metrics_path": "/v1/metrics?format=prometheus",
				"file_sd_configs": []map[string][]string{
					{"files": {"/model_bazaar/cluster-monitoring/nomad_nodes/*.yaml"}},
				},
				"relabel_configs": []map[string]interface{}{
					{
						"source_labels": []string{"__address__"},
						"regex":         "([^:]+):.+",
						"target_label":  "hostname",
						"replacement":   "nomad-agent-${1}",
					},
				},
			},
		}
	} else {
		orchestratorEntry = []map[string]interface{}{
			{
				"job_name": "kubelet",
				"kubernetes_sd_configs": []map[string]interface{}{
					{"role": "node"},
				},
				"scheme": "https",
				"tls_config": map[string]interface{}{
					"ca_file": "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				},
				"bearer_token_file": "/var/run/secrets/kubernetes.io/serviceaccount/token",
				"metrics_path":      "/metrics",
				"relabel_configs": []map[string]interface{}{
					{"action": "labelmap", "regex": "__meta_kubernetes_node_label_(.+)"},
				},
				"metric_relabel_configs": []map[string]interface{}{
					{
						"action": "labeldrop",
						"regex":  "(beta_kubernetes_io_.*|topology_.*|eks_amazonaws_com_.*)",
					},
				},
			},
			{
				"job_name": "kubelet-cadvisor",
				"kubernetes_sd_configs": []map[string]interface{}{
					{"role": "node"},
				},
				"scheme": "https",
				"tls_config": map[string]interface{}{
					"ca_file": "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				},
				"bearer_token_file": "/var/run/secrets/kubernetes.io/serviceaccount/token",
				"metrics_path":      "/metrics/cadvisor",
				"relabel_configs": []map[string]string{
					{
						"action": "labelmap",
						"regex":  "__meta_kubernetes_node_label_(.+)",
					},
				},
				"metric_relabel_configs": []map[string]interface{}{
					{
						"action": "labeldrop",
						"regex":  "(beta_kubernetes_io_.*|topology_.*|eks_amazonaws_com_.*)",
					},
				},
			},
			{
				"job_name": "kube-state-metrics",
				"kubernetes_sd_configs": []map[string]interface{}{
					{"role": "service", "namespaces": map[string]interface{}{"names": []string{"kube-system"}}},
				},
				"relabel_configs": []map[string]interface{}{
					{"source_labels": []string{"__meta_kubernetes_service_name"}, "action": "keep", "regex": "kube-state-metrics"},
				},
			},
			{
				"job_name": "node-exporter",
				"kubernetes_sd_configs": []map[string]interface{}{
					{
						"role": "endpoints",
						"namespaces": map[string]interface{}{
							"names": []string{"kube-system"},
						},
					},
				},
				"relabel_configs": []map[string]interface{}{
					{
						"source_labels": []string{"__meta_kubernetes_service_name"},
						"regex":         "node-exporter-prometheus-node-exporter",
						"action":        "keep",
					},
					{
						"source_labels": []string{"__meta_kubernetes_endpoint_node_name"},
						"target_label":  "node",
						"action":        "replace",
					},
				},
				"metric_relabel_configs": []map[string]interface{}{
					{
						"action": "labeldrop",
						"regex":  "(beta_kubernetes_io_.*|topology_.*|eks_amazonaws_com_.*)",
					},
				},
			},
		}
	}

	return map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "1s",
		},
		"scrape_configs": append(orchestratorEntry, map[string]interface{}{
			"job_name":        "deployment-jobs",
			"metrics_path":    "/metrics",
			"http_sd_configs": []map[string]string{{"url": deploymentTargetsEndpoint}},
			"relabel_configs": []map[string]interface{}{
				{
					"source_labels": []string{"model_id"},
					"target_label":  "workload",
					"replacement":   "deployment-${1}",
				},
			},
		}),
	}
}

func createVectorConfig(storage storage.Storage, modelBazaarEndpoint string) error {
	config := map[string]interface{}{
		// the checkpoints for different logs will be stored in this directory
		"data_dir": "/model_bazaar/logs",
		"sources": map[string]interface{}{
			// fetching logs for all models
			"training_logs": map[string]interface{}{
				"type": "file",
				"include": []string{
					"/model_bazaar/logs/*/train.log",
				},
				"read_from": "beginning",
			},
			"deployment_logs": map[string]interface{}{
				"type": "file",
				"include": []string{
					"/model_bazaar/logs/*/deployment.log",
				},
				"read_from": "beginning",
			},
		},
		"transforms": map[string]interface{}{
			"parse_logs": map[string]interface{}{
				"type": "remap",
				"inputs": []string{
					"training_logs",
					"deployment_logs",
				},
				"source": ". = parse_json!(.message)",
			},
			"filter_debug_logs": map[string]interface{}{
				"type": "filter",
				"inputs": []string{
					"parse_logs",
				},
				"condition": ".level == \"DEBUG\"",
			},
			"filter_non_debug_logs": map[string]interface{}{
				"type": "filter",
				"inputs": []string{
					"parse_logs",
				},
				"condition": ".level != \"DEBUG\"",
			},
		},
		"sinks": map[string]interface{}{
			"debug_logs": map[string]interface{}{
				"type": "http",
				"inputs": []string{
					"filter_debug_logs",
				},
				"uri": strings.TrimSuffix(modelBazaarEndpoint, "/") + "/victorialogs/insert/jsonline?_stream_fields=model_id,service_type&_msg_field=_msg&_time_field=_time&extra_fields=retention=48h",
				"encoding": map[string]interface{}{
					"codec": "json",
				},
				"framing": map[string]interface{}{
					"method": "newline_delimited",
				},
				"request": map[string]interface{}{
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
					},
				},
				"healthcheck": map[string]interface{}{
					"enabled": false,
				},
			},
			"other_logs": map[string]interface{}{
				"type": "http",
				"inputs": []string{
					"filter_non_debug_logs",
				},
				"uri": strings.TrimSuffix(modelBazaarEndpoint, "/") + "/victorialogs/insert/jsonline?_stream_fields=model_id,service_type&_msg_field=_msg&_time_field=_time&extra_fields=retention=30d",
				"encoding": map[string]interface{}{
					"codec": "json",
				},
				"framing": map[string]interface{}{
					"method": "newline_delimited",
				},
				"request": map[string]interface{}{
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
					},
				},
				"healthcheck": map[string]interface{}{
					"enabled": false,
				},
			},
		},
	}

	configFile, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error creating vector config: %w", err)
	}

	err = storage.Write(
		filepath.Join("cluster-monitoring", "vector", "vector.yaml"),
		bytes.NewReader(configFile),
	)
	if err != nil {
		return fmt.Errorf("error writing vector config: %w", err)
	}

	return nil
}

func createGrafanaProvisionings(storage storage.Storage, isLocal bool, orchestratorName string, modelBazaarEndpoint string) error {
	// Create grafana dashboard config
	dashboardConfig := map[string]interface{}{
		"apiVersion": 1,
		"providers": []map[string]interface{}{
			{
				"name":                  "dashboards",
				"type":                  "file",
				"disableDeletion":       true,
				"updateIntervalSeconds": 10,
				"allowUiUpdates":        true,
				"options": map[string]interface{}{
					"foldersFromFilesStructure": true,
					"path":                      filepath.Join(storage.Location(), "cluster-monitoring", "grafana", "dashboards"),
				},
			},
		},
	}
	configFile, err := yaml.Marshal(dashboardConfig)
	if err != nil {
		return fmt.Errorf("error creating grafana dashboard config: %w", err)
	}

	err = storage.Write(
		filepath.Join("cluster-monitoring", "grafana", "provisioning", "dashboards", "dashboards.yaml"),
		bytes.NewReader(configFile),
	)
	if err != nil {
		return fmt.Errorf("error writing grafana dashboards config: %w", err)
	}

	// Create grafana datasource config
	var url string
	if isLocal && orchestratorName == "nomad" {
		url = "http://host.docker.internal/victoriametrics"
	} else {
		url = strings.TrimSuffix(modelBazaarEndpoint, "/") + "/victoriametrics"
	}
	datasourceConfig := map[string]interface{}{
		"apiVersion": 1,
		"datasources": []map[string]interface{}{
			{
				"name":      "Prometheus",
				"type":      "prometheus",
				"url":       url,
				"access":    "proxy",
				"isDefault": true,
				"editable":  false,
			},
		},
	}
	configFile, err = yaml.Marshal(datasourceConfig)
	if err != nil {
		return fmt.Errorf("error creating grafana datasource config: %w", err)
	}

	err = storage.Write(
		filepath.Join("cluster-monitoring", "grafana", "provisioning", "datasources", "datasources.yaml"),
		bytes.NewReader(configFile),
	)
	if err != nil {
		return fmt.Errorf("error writing grafana datasources config: %w", err)
	}

	return nil
}
