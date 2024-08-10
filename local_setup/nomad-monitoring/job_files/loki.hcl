job "loki" {
  datacenters = ["dc1"]
  type        = "service"

  group "loki" {

    network {
      port "loki_port" {
        to = 3100
      }
    }
        
    count = 1

    task "loki" {
      driver = "docker"

      config {
        image = "grafana/loki:3.0.0"
        args = [
          "-config.file",
          "$${NOMAD_TASK_DIR}/loki.yaml",
        ]

        volumes = [
            "/home/gautam/ThirdAI-Platform/local_setup/nomad-monitoring/data/loki:/loki_data"
        ]

        ports = ["loki_port"]
      }

      resources {
        cpu    = 5000
        memory = 2000
      }

      template {
        data            = file(abspath("./../configs/loki.yaml"))
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