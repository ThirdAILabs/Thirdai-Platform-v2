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

# This means jobs would be querying every 5m, for the key
# if a key is set it would be available in the environment
# only after 5m of being set
vault {
  enabled = true
  address = "http://127.0.0.1:8200"
  task_token_ttl = "5m"
  create_from_role = "nomad-cluster"
  token = "00000000-0000-0000-0000-000000000000"
}