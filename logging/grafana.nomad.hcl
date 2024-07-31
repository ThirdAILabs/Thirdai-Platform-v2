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
      }

      config {
        image = "grafana/grafana"

        ports = ["grafana_port"]

        volumes = [
          "/Users/kartiksarangmath/Documents/thirdai/ThirdAI-Platform/logging/grafana_datasource.yaml:/etc/grafana/provisioning/datasources/datasource.yaml"
        ]

      }

      resources {
        cpu    = 5000
        memory = 5000
      }

      service {
        name = "graphana"
        port = "grafana_port"
        provider = "nomad"
      }
    }

    network {
          port  "grafana_port"{
            to = 3000
          }
        }

  }
}