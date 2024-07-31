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
          "/etc/loki/loki.yaml",  // Adjusted to reflect the mounted path
        ]

        volumes = [
          "/Users/kartiksarangmath/Documents/thirdai/ThirdAI-Platform/logging/loki.yaml:/etc/loki/loki.yaml"
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
      }
    }
  }
}