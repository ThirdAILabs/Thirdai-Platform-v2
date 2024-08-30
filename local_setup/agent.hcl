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
    extra_labels             = ["job_name", "job_id", "task_group_name", "task_name", "namespace", "node_name", "node_id"]
  }
}

plugin "raw_exec" {
  config {
    enabled = true
  }
}

telemetry {
  collection_interval        = "2s"
  disable_hostname           = true
  prometheus_metrics         = true
  publish_allocation_metrics = true
  publish_node_metrics       = true
}

limits {
  http_max_conns_per_client = 0
  rpc_max_conns_per_client = 0
}