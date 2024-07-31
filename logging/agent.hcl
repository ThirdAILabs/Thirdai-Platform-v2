# nomad-agent.hcl

# Specify the bind address for the agent
bind_addr = "0.0.0.0"

# Configure the HTTP interface
addresses {
  http = "0.0.0.0"
}

# Set the HTTP port to 4777
ports {
  http = 4777
  rpc  = 4778
  serf = 4779
}
# Enable the client mode
client {
  enabled = true
  host_volume "loki-data" {
    path      = "/home/pratyush/ThirdAI-Platform/logging/docker-promtail-loki/loki-data"
    read_only = false
}

host_volume "grafana-data" {
    path      = "/home/pratyush/ThirdAI-Platform/logging/docker-promtail-loki/grafana-data"
    read_only = false
}
}

# Enable the server mode (for dev setup, you might want both client and server)
server {
  enabled = true
  bootstrap_expect = 1
}

# Set the data directory
data_dir = "/home/pratyush/ThirdAI-Platform/logging/docker-promtail-loki/data"

# Adjust log level if needed
log_level = "INFO"

# Enable the UI
ui {
  enabled = true
}