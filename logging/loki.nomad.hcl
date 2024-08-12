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
          "/etc/loki/loki.yaml",
        ]

        volumes = [
          "/Users/benitogeordie/ThirdAI-Platform-2/logging/loki.yaml:/etc/loki/loki.yaml",
          "/Users/benitogeordie/ThirdAI-Platform-2/logging/loki_data:/loki_data"  # ensure that this directory has permissions for this docker container user to do everything
        ]

        ports = ["loki_port"]
      }

      resources {
        cpu    = 5000
        memory = 2000
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