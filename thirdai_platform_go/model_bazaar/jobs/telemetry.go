package jobs

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"thirdai_platform/model_bazaar/nomad"
	"thirdai_platform/model_bazaar/storage"

	"gopkg.in/yaml.v3"
)

type TelemetryJobArgs struct {
	IsLocal             bool
	ModelBazaarEndpoint string
	Docker              nomad.DockerEnv
	GrafanaDbUrl        string

	AdminUsername string
	AdminEmail    string
	AdminPassword string
}

func StartTelemetryJob(client nomad.NomadClient, storage storage.Storage, args TelemetryJobArgs) error {
	slog.Info("starting telemetry job")

	err := createPromFile(storage, args.ModelBazaarEndpoint, args.IsLocal)
	if err != nil {
		return fmt.Errorf("error creating promfile: %w", err)
	}

	url, err := url.Parse(args.ModelBazaarEndpoint)
	if err != nil {
		return fmt.Errorf("unable to parse model bazaar endpoint: %w", err)
	}

	targetCount, err := countTargets(storage)
	if err != nil {
		return fmt.Errorf("error counting telemetry targets: %w", err)
	}

	// TODO: how to ensure grafana dashboards in in expected location

	job := nomad.TelemetryJob{
		IsLocal:                args.IsLocal,
		TargetCount:            targetCount,
		NomadMonitoringDir:     "/model_bazaar/nomad-monitoring",
		AdminUsername:          args.AdminUsername,
		AdminEmail:             args.AdminEmail,
		AdminPassword:          args.AdminPassword,
		GrafanaDbUrl:           args.GrafanaDbUrl,
		ModelBazaarPrivateHost: url.Hostname(),
		Docker:                 args.Docker,
	}

	err = stopJobIfExists(client, job.GetJobName())
	if err != nil {
		slog.Error("error stopping existing telemetry job", "error", err)
		return fmt.Errorf("error stopping existing telemetry job: %w", err)
	}

	err = client.StartJob(job)
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

func createPromFile(storage storage.Storage, modelBazaarEndpoint string, isLocal bool) error {
	serverNodeFile := filepath.Join("nomad-monitoring", "nomad_nodes", "server.yaml")

	if isLocal {
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

	promfile, err := yaml.Marshal(prometheusConfig(modelBazaarEndpoint, isLocal))
	if err != nil {
		return fmt.Errorf("error creating promfile: %w", err)
	}

	err = storage.Write(
		filepath.Join("nomad-monitoring", "node_discovery", "prometheus.yaml"),
		bytes.NewReader(promfile),
	)
	if err != nil {
		return fmt.Errorf("error writing promfile: %w", err)
	}

	return nil
}

func getDeploymentTargetsEndpoint(modelBazaarEndpoint string, isLocal bool) string {
	if isLocal {
		return "http://host.docker.internal:80/api/telemetry/deployment-services"
	} else {
		return strings.TrimSuffix(modelBazaarEndpoint, "/") + "/api/telemetry/deployment-services"
	}
}

func prometheusConfig(modelBazaarEndpoint string, isLocal bool) map[string]interface{} {
	deploymentTargetsEndpoint := getDeploymentTargetsEndpoint(modelBazaarEndpoint, isLocal)

	return map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "1s",
			"external_labels": map[string]string{"env": "dev", "cluster": "local"},
		},
		"scrape_configs": []map[string]interface{}{
			{
				"job_name":     "nomad-agent",
				"metrics_path": "/v1/metrics?format=prometheus",
				"file_sd_configs": []map[string][]string{
					{"files": {"/model_bazaar/nomad-monitoring/nomad_nodes/*.yaml"}},
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
			{
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
			},
		},
	}
}

func countTargets(storage storage.Storage) (int, error) {
	count := 0
	for _, nodeType := range []string{"server.yaml", "client.yaml"} {
		nodeFile := filepath.Join("nomad-monitoring", "nomad_nodes", nodeType)

		exists, err := storage.Exists(nodeFile)
		if err != nil {
			return -1, fmt.Errorf("error counting targets while checking if %v exists: %w", nodeFile, err)
		}

		if exists {
			file, err := storage.Read(nodeFile)
			if err != nil {
				return -1, fmt.Errorf("error counting targets while reading %v: %w", nodeFile, err)
			}
			defer file.Close()

			var targets []targetList
			err = yaml.NewDecoder(file).Decode(&targets)
			if err != nil {
				return -1, fmt.Errorf("error counting targets while parsing %v: %w", nodeFile, err)
			}

			for _, targetList := range targets {
				count += len(targetList.Targets)
			}
		}
	}

	return count, nil
}
