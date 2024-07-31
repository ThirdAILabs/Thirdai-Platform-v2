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
          "/etc/loki/loki.yaml",  // Adjusted to reflect the mounted path, you need to ensure that the user running this docker container has access to write to this directory
        ]

        volumes = [
          "/home/kartik/ThirdAI-Platform/logging/loki.yaml:/etc/loki/loki.yaml",
          "/home/kartik/ThirdAI-Platform/logging/loki-data:/loki-data"  # make sure to create the directory "loki-data"
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