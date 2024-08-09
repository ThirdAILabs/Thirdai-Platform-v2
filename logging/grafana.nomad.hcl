job "grafana" {
  datacenters = ["dc1"]
  type        = "service"

  group "grafana" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "grafana" {
      driver = "docker"

      env {
        GF_AUTH_ANONYMOUS_ENABLED = "true"
        GF_AUTH_BASIC_ENABLED = "false"
        GF_LOG_LEVEL = "DEBUG"
        GF_LOG_MODE = "console"
        GF_AUTH_ANONYMOUS_ORG_ROLE = "Admin"
        GF_AUTH_DISABLE_LOGIN_FORM = "true"
        GF_SERVER_ROOT_URL = "%(protocol)s://%(domain)s:%(http_port)s/grafana/"
        GF_SERVER_SERVE_FROM_SUB_PATH = "true"
      }

      config {
        image = "grafana/grafana"

        ports = ["grafana_port"]

        volumes = [
          "/home/kartik/ThirdAI-Platform/logging/grafana_datasource.yaml:/etc/grafana/provisioning/datasources/datasource.yaml"
        ]

      }

      resources {
        cpu    = 5000
        memory = 5000
      }

      service {
        name = "grafana"
        port = "grafana_port"
        provider = "nomad"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.grafana_port.rule=PathPrefix(`/grafana`)",
          "traefik.http.routers.grafana_port.priority=10"
        ]
      }
    }

    network {
          port  "grafana_port"{
            to = 3000
          }
        }

  }
}