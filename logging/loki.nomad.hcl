job "loki" {
  datacenters = ["dc1"]
  type        = "service"

  group "loki" {

    network {
      port "loki_port" {
        static = 3100
      }
    }
        
    count = 1

    volume "loki-config" {
      type      = "host"
      source    = "loki-config"
      read_only = false
    }

    task "loki" {
      driver = "docker"

      volume_mount {
        volume      = "loki-config"
        destination = "/etc/loki"
        read_only   = false
      }

      config {
        image = "grafana/loki:3.0.0"

        args = [
          "-config.file",
          "/etc/loki/loki.yaml",  // Adjusted to reflect the mounted path
        ]

        ports = ["loki_port"]
      }

      resources {
        cpu    = 5000
        memory = 3200
      }

      service {
        name = "loki"
        port = "loki_port"

        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}