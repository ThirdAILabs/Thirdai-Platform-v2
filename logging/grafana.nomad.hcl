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

      config {
        image = "grafana/grafana"

        ports = ["grafana_port"]
      }

      resources {
        cpu    = 5000
        memory = 32000
      }

      service {
        name = "graphana"
        port = "grafana_port"

        check {
          type     = "http"
          path     = "/health"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }

    network {
          port  "grafana_port"{
            static = 3000
            to = 3000
          }
        }

  }
}