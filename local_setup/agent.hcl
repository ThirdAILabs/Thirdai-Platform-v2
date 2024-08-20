bind_addr = "0.0.0.0"
data_dir  = "/opt/nomad/data"

client {
  enabled           = true
}

server {
  enabled              = true
  bootstrap_expect = 1
}

plugin "docker" {
  config {
    volumes {
      enabled = true
    }
  }
}