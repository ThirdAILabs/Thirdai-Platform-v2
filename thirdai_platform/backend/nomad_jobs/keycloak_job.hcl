job "keycloak" {
  datacenters = ["dc1"]

  group "keycloak-group" {
    count = 1

    network {
      port "keycloak-http" {
        static = 8180
      }
    }

    task "keycloak" {
      driver = "docker"

      config {
        image = "quay.io/keycloak/keycloak:22.0"
        ports = ["keycloak-http"]
        args  = ["start", "--http-port=8180"]
      }

      env {
        KC_HEALTH_ENABLED            = "true"
        KC_METRICS_ENABLED           = "true"
        KC_HTTP_ENABLED              = "true"
        KC_HOSTNAME_STRICT_HTTPS      = "false"
        KEYCLOAK_SSL_REQUIRED        = "none"
        KC_HOSTNAME_STRICT_BACKCHANNEL = "false"
        KC_PROXY                     = "edge"
        KC_HOSTNAME_URL              = "http://localhost/keycloak"

        # Database connection
        DB_VENDOR                    = "postgres"
        DB_ADDR                      = "172.17.0.1"  # Using Docker bridge gateway IP
        DB_DATABASE                  = "keycloakdb"
        DB_USER                      = "postgres"
        DB_PASSWORD                  = "newpassword"

        # Keycloak DB configuration
        KC_DB                        = "postgres"
        KC_DB_URL                    = "jdbc:postgresql://172.17.0.1:5432/keycloakdb"  # Using Docker bridge gateway IP
        KC_DB_USERNAME               = "postgres"
        KC_DB_PASSWORD               = "newpassword"
      }

      resources {
        cpu    = 500
        memory = 512
      }

      service {
        name = "keycloak"
        port = "keycloak-http"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.keycloak-http.rule=PathPrefix(`/keycloak/`)",
          "traefik.http.routers.keycloak-http.priority=10",
          "traefik.http.services.keycloak-http.loadbalancer.server.port=8180"
        ]
        provider = "nomad"
      }
    }
  }
}
