job "victoria-magnet" {
  datacenters = ["dc1"]
  type        = "service"

  group "victoria-magent-group" {
    count = 1

    volume "victoriametrics" {
      type      = "host"
      read_only = false
      source    = "victoriametrics"
    }

    network {
      mode = "bridge"
      port "vicky-http" {
        static = 8248
        to = 8428
      }

      port "vmagent-http" {
        static = 8429
        to = 8429
      }
    }

    task "victoriametrics" {
      lifecycle {
        hook    = "prestart"
        sidecar = true
      }
      driver = "docker"

      service {
        name     = "vicky-web"
        provider = "nomad"
        tags     = ["victoriametrics", "web"]
        port     = "vicky-http"
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
          "--retentionPeriod=1",
          "--httpListenAddr=:8428"
        ]
      }

      resources {
        cpu    = 256
        memory = 300
      }
    }

    task "vmagent" {
      driver = "docker"

      config {
        image = "victoriametrics/vmagent:latest"
        args = [
          "--promscrape.config=$${NOMAD_TASK_DIR}/prometheus.yml",
          "--remoteWrite.url=${VICTORIAMETRICS_ADDR}"
        ]
      }

      template {
        data        = file(abspath("./configs/prometheus.tpl.yml"))
        destination = "$${NOMAD_TASK_DIR}/prometheus.yml"
        change_mode = "restart"
      }

      template {
        data = <<EOF
          {{- range nomadService "vicky-web" }}
          VICTORIAMETRICS_ADDR=http://{{ .Address }}:{{ .Port }}/api/v1/write
        {{ end -}}
        EOF

        destination = "local/env"
        env         = true
      }


      resources {
        cpu    = 256
        memory = 300
      }
    }
  }

}