bind_addr = "0.0.0.0"
data_dir  = "/opt/nomad/data"

client {
  enabled                     = true
}

server {
  enabled                    = true
}

plugin "docker" {
  config {
    volumes {
      enabled                = true
    }
  }
}

limits {
  http_max_conns_per_client = 0
  rpc_max_conns_per_client = 0
}