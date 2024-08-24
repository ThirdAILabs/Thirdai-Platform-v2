job "redis-server" {
  datacenters = ["dc1"]

  group "redis" {
    count = 1

    network {
      port "db" {
        static = 6379
      }
    }

    task "redis" {
      driver = "docker"

      config {
        image = "redis:latest"
        ports = ["db"]
      }

      resources {
        cpu    = 500
        memory = 256
      }

      service {
        name = "redis"
        provider = "nomad"
        port = "db"
        tags = ["traefik.enable=true", "traefik.http.services.redis.loadbalancer.server.port=6379"]

        check {
          type     = "tcp"
          port     = "db"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}
