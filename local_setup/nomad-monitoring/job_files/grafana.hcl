job "grafana" {
  datacenters = ["dc1"]
  type        = "service"

  group "grafana" {
    count = 1

    volume "grafana" {
      type      = "host"
      read_only = false
      source    = "grafana"
    }

    network {
      mode = "bridge"

      port "grafana-http" {
        static = 3000
        to = 3000
      }
    }

    task "grafana" {
      driver = "docker"

      service {
        name     = "grafana-web"
        port     = "grafana-http"
        provider = "nomad"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.grafana_port.rule=PathPrefix(`/grafana`)",
          "traefik.http.routers.grafana_port.priority=10"
        ]
      }

      env {
        GF_AUTH_ANONYMOUS_ENABLED = "true"
        GF_AUTH_BASIC_ENABLED = "false"
        GF_LOG_LEVEL          = "DEBUG"
        GF_LOG_MODE           = "console"
        GF_AUTH_ANONYMOUS_ORG_ROLE = "Admin"
        GF_AUTH_DISABLE_LOGIN_FORM = "true"
        GF_SERVER_ROOT_URL = "%(protocol)s://%(domain)s:%(http_port)s/grafana/"
        GF_SERVER_SERVE_FROM_SUB_PATH = "true"
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
        image = "grafana/grafana:main-ubuntu"
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
    {{ range nomadService "vicky-web" }}
    url: http://{{ .Address }}:{{ .Port }}
    {{ end }}
  - name: Loki
    type: loki
    access: proxy
    {{ range nomadService "vicky-web" }}
    url: http://{{ .Address }}
    {{ end }}
    version: 1
    editable: false
    isDefault: true
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