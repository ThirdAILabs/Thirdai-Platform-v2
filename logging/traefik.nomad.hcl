job "traefik" {
  datacenters = ["dc1"]

  type = "service"

  group "traefik" {
    count = 1

    network {
      port "http" {
        static = 80
      }
      port "admin" {
        static = 8080
      }
    }

    service {
      name = "traefik-http"
      provider = "nomad"
      port = "http"
    }

    task "server" {
      driver = "docker"

      config {
        image = "traefik:v2.10"
        ports = ["admin", "http"]
        args = [
          "--api.dashboard=true",
          "--entrypoints.web.address=:${NOMAD_PORT_http}",
          "--entrypoints.traefik.address=:${NOMAD_PORT_admin}",
          "--providers.nomad=true",
          "--providers.nomad.endpoint.address=http://192.168.1.11:4646", ### IP to your nomad server 
          "--entrypoints.websecure.http.tls=true",
          "--log.level=DEBUG"
        ]
      }
    }
  }
}