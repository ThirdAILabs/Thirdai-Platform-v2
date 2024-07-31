bind_addr = "0.0.0.0"
data_dir  = "/opt/nomad/data"

client {
  enabled           = true
}

server {
  enabled              = true
}

plugin "docker" {
  config {
    volumes {
      enabled = true
    }
  }
}

client {
  enabled = true
  host_volume "loki-config" {
    path      = "../logging/docker-promtail-loki/config/loki"
    read_only = false
  }
}
