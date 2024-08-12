bind_addr = "0.0.0.0"
data_dir  = "/Users/benitogeordie/nomad/data"

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
    endpoint = "unix:///Users/benitogeordie/.docker/run/docker.sock"
  }
}