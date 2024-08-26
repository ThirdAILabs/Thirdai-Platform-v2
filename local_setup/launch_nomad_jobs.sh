nomad job run -var="nomad_endpoint=$(nomad agent-info | grep 'known_servers' | awk '{print $3}' | sed 's/:4647//')" autoscaler.nomad

nomad job run redis.nomad