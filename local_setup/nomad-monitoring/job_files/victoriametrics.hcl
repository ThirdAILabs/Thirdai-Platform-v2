job "victoriametrics" {
  datacenters = ["dc1"]
  type        = "service"

  group "victoriametrics" {
    count = 1

    volume "victoriametrics" {
      type      = "host"
      read_only = false
      source    = "victoriametrics"
    }

    network {
      mode = "bridge"

      port "vicky-http" {
        static = 8428
        to = 8428
      }
    }

    task "victoriametrics" {
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
          "--httpListenAddr=:8428",
          "--promscrape.config=$${NOMAD_TASK_DIR}/prometheus.yml"
        ]
      }

      template {
        data        = file(abspath("./../configs/prometheus.tpl.yml"))
        destination = "$${NOMAD_TASK_DIR}/prometheus.yml"
        change_mode = "restart"
      }

      resources {
        cpu    = 256
        memory = 300
      }
    }
  }

}