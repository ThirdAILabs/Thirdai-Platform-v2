job "victoria-loki" {
  datacenters = ["dc1"]
  type        = "service"

  group "victoria-loki" {
    count = 1

    volume "victoriametrics" {
      type      = "host"
      read_only = false
      source    = "victoriametrics"
    }

    volume "loki" {
      type      = "host"
      read_only = false
      source    = "loki"
    }

    network {
      mode = "bridge"

      port "vicky-http" {
        to = 8428
      }

      port "loki_port" {
        to = 3100
      }
    }

    task "preprocess" {
      driver = "exec"
      lifecycle {
        hook = "prestart"
      }
      config {
        command = "/bin/sh"
        args = ["-c", "rm -rf /storage/* /loki_data/*"]
      }
      volume_mount {
        volume      = "victoriametrics"
        destination = "/storage"
        read_only   = false
      }

      volume_mount {
        volume      = "loki"
        destination = "/loki_data"
        read_only   = false
      }
    }

    task "victoriametrics" {
      driver = "docker"

      service {
        name     = "vicky-web"
        provider = "nomad"
        port     = "vicky-http"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.victoriametric-http.rule=PathPrefix(`/victoria-metric`)",
          "traefik.http.routers.victoriametric-http.priority=10"
        ]
      }

      volume_mount {
        volume      = "victoriametrics"
        destination = "/storage"
        read_only   = false
      }

      config {
        image = "victoriametrics/victoria-metrics:latest"
        ports = ["vicky-http"]
        args = [
          "--storageDataPath=/storage",
          "--retentionPeriod=1d",
          "--httpListenAddr=:${NOMAD_PORT_vicky_http}",
          "--promscrape.config=$${NOMAD_TASK_DIR}/prometheus.yaml"
        ]
      }

      template {
        data        = <<EOF
global:
  scrape_interval: 2s
  external_labels:
    env: "dev"
    cluster: "local"

scrape_configs:
  - job_name: "nomad-agent"
    metrics_path: "/v1/metrics?format=prometheus"
    static_configs:
      - targets: {{ range nomadService "vicky-web" }} ["{{ .Address }}:4646"] {{ end }}
        labels:
          role: agent
    relabel_configs:
      - source_labels: [__address__]
        regex: "([^:]+):.+"
        target_label: "hostname"
        replacement: "nomad-agent-$1"
EOF
        destination = "$${NOMAD_TASK_DIR}/prometheus.yaml"
        change_mode = "restart"
      }

      resources {
        cpu    = 256
        memory = 3000
      }
    }

    task "loki" {
      driver = "docker"

      volume_mount {
        volume      = "loki"
        destination = "/loki_data"
        read_only   = false
      }

      user = "root"

      config {
        image = "grafana/loki:3.0.0"
        args = [
          "-config.file",
          "$${NOMAD_TASK_DIR}/loki.yaml",
        ]

        ports = ["loki_port"]
      }

      resources {
        cpu    = 5000
        memory = 2000
      }

      /*
        Date: 19/08/2024
        Cannot use "data            = file(abspath("./../configs/loki.yaml"))" because file function only works via CLI currently.
        https://github.com/hashicorp/nomad/issues/19648#issuecomment-1881052624
        */
      template {
        data            = <<EOF
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  instance_addr: 127.0.0.1
  path_prefix: /loki_data
  storage:
    filesystem:
      chunks_directory: /loki_data/chunks
      rules_directory: /loki_data/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

frontend:
  max_outstanding_per_tenant: 2048

pattern_ingester:
  enabled: true

limits_config:
  max_global_streams_per_user: 0
  ingestion_rate_mb: 50000
  ingestion_burst_size_mb: 50000
  volume_enabled: true

query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 100

schema_config:
  configs:
    - from: 2020-10-27
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://localhost:9093

analytics:
  reporting_enabled: false

ingester:
  wal:
    flush_on_shutdown: true
EOF
        destination     = "$${NOMAD_TASK_DIR}/loki.yaml"
      }

      service {
        name = "loki"
        port = "loki_port"
        provider = "nomad"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.modelbazaar-http.rule=PathPrefix(`/loki`)",
          "traefik.http.routers.modelbazaar-http.priority=10"
        ]
      }
    }
  }
}