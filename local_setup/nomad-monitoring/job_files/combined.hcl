job "combined" {
  datacenters = ["dc1"]
  type        = "service"

  group "victoriametrics" {
    count = 1

    volume "victoriametrics" {
      type      = "host"
      read_only = false
      source    = "victoriametrics"
    }

    volume "grafana" {
      type      = "host"
      read_only = false
      source    = "grafana"
    }

    network {
      mode = "bridge"

      port "vicky-http" {
        static = 8428
        to = 8428
      }
      port "vmagent-http" {
        static = 8429
        to = 8429
      }

      port "grafana-http" {
        static = 3000
        to = 3000
      }
    }

    task "victoriametrics" {
      lifecycle{
        hook = "prestart"
        sidecar = true
      }

      driver = "docker"

      service {
        name     = "vicky-web"
        provider = "nomad"
        port     = "vicky-http"
        tags = [
          "victoriametrics",
          "web",
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
          "--httpListenAddr=:8428"
        ]
      }

      resources {
        cpu    = 256
        memory = 300
      }
    }

    task "vmagent" {
      driver = "docker"

      config {
        image = "victoriametrics/vmagent:latest"
        args = [
          "--promscrape.config=$${NOMAD_TASK_DIR}/prometheus.yml",
          "--remoteWrite.url=${VICTORIAMETRICS_ADDR}"
        ]
      }

      template {
        data        = file(abspath("./../configs/prometheus.tpl.yml"))
        destination = "$${NOMAD_TASK_DIR}/prometheus.yml"
        change_mode = "restart"
      }

      template {
        data = <<EOF
{{- range nomadService "vicky-web" }}
  VICTORIAMETRICS_ADDR=http://{{ .Address }}:{{ .Port }}/api/v1/write
{{ end -}}
EOF

        destination = "local/env"
        env         = true
      }


      resources {
        cpu    = 256
        memory = 300
      }
    }

    task "grafana" {

      lifecycle{
        hook = "poststart"
        sidecar = true
      }
      driver = "docker"

      service {
        name     = "grafana-web"
        port     = "grafana-http"
        provider = "nomad"
        tags = [
          "grafana",
          "web",
          "traefik.enable=true",
          "traefik.http.routers.grafana_port.rule=PathPrefix(`/grafana`)",
          "traefik.http.routers.grafana_port.priority=10"
        ]
      }

      env {
        GF_LOG_LEVEL          = "DEBUG"
        GF_LOG_MODE           = "console"
        GF_SERVER_HTTP_PORT   = "$${NOMAD_PORT_http}"
        GF_PATHS_PROVISIONING = "/local/grafana/provisioning"
      }

      volume_mount {
        volume      = "grafana"
        destination = "/var/lib/grafana"
        read_only   = false
      }

      user = "root"

      config {
        image = "grafana/grafana:9.1.4-ubuntu"
        ports = ["grafana-http"]
      }

      resources {
        cpu    = 256
        memory = 300
      }

      template {
        data        = <<EOF
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    {{- range nomadService "vicky-web" }}
    url: http://{{.Address}}:{{.Port}}
    {{ end -}}
EOF
        destination = "/local/grafana/provisioning/datasources/datasources.yaml"
      }

      template {
        data        = <<EOF
apiVersion: 1
providers:
  - name: dashboards
    type: file
    updateIntervalSeconds: 10
    options:
      foldersFromFilesStructure: true
      path: /local/grafana/provisioning/dashboards
EOF
        destination = "/local/grafana/provisioning/dashboards/dashboards.yaml"
      }

      template {
        data            = file(abspath("./../dashboards/allocations.json"))
        destination     = "local/grafana/provisioning/dashboards/nomad/allocations.json"
        left_delimiter  = "[["
        right_delimiter = "]]"
      }
      template {
        data            = file(abspath("./../dashboards/clients.json"))
        destination     = "local/grafana/provisioning/dashboards/nomad/clients.json"
        left_delimiter  = "[["
        right_delimiter = "]]"
      }
      template {
        data            = file(abspath("./../dashboards/server.json"))
        destination     = "local/grafana/provisioning/dashboards/nomad/server.json"
        left_delimiter  = "[["
        right_delimiter = "]]"
      }
    }

  }

}