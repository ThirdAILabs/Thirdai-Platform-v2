bind_addr = "0.0.0.0"
data_dir  = "/opt/nomad/data"

client {
  enabled                     = true

  meta {
    env = "dev"
  }
  
  host_volume "grafana" {
    path                      = "/opt/nomad/nomad-monitoring/data/grafana"
    read_only                 = false
  }

  host_volume "victoriametrics" {
    path                      = "/opt/nomad/nomad-monitoring/data/victoriametric"
    read_only                 = false
  }
}

server {
  enabled                    = true
}

plugin "docker" {
  config {
    allow_privileged = true
    volumes {
      enabled                = true
    }
    extra_labels             = ["job_name", "job_id", "task_group_name", "task_name", "namespace", "node_name", "node_id"]
  }
}

telemetry {
  collection_interval        = "15s"
  disable_hostname           = true
  prometheus_metrics         = true
  publish_allocation_metrics = true
  publish_node_metrics       = true
}

limits {
  http_max_conns_per_client = 0
  rpc_max_conns_per_client = 0
}