bind_addr = "0.0.0.0"
data_dir  = "/opt/nomad/data"

advertise {
  http = "192.168.1.8"
  rpc  = "192.168.1.8"
  serf = "192.168.1.8"
}

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

limits {
  http_max_conns_per_client = 0
  rpc_max_conns_per_client = 0
}