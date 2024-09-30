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
        network_mode = "host"
        image = "quay.io/keycloak/keycloak:22.0"
        ports = ["keycloak-http"]
        args = [
          "start",
          "--auto-build",
          "--hostname-strict=false",
          "--hostname-strict-https=false",
          "--http-enabled=true",
          "--metrics-enabled=true",
          "--db=postgres",
          "--health-enabled=true",
        ]
      }

      env {
          # KC_HEALTH_ENABLED            = "true"
          # KC_METRICS_ENABLED           = "true"
          # KC_HTTP_ENABLED              = "true"
          # KC_HOSTNAME_STRICT_HTTPS      = "false"
          # KEYCLOAK_SSL_REQUIRED        = "none"
          # KC_HOSTNAME_STRICT_BACKCHANNEL = "false"
          KC_HOSTNAME                  = "localhost"
          KC_HOSTNAME_PORT             = "8180"

          # Database connection
          # DB_VENDOR                    = "postgres"
          # DB_ADDR                      = "localhost"  # No change needed, localhost should work in host network
          # DB_DATABASE                  = "keycloakdb"
          # DB_USER                      = "postgres"
          # DB_PASSWORD                  = "newpassword"

          # Keycloak Admin credentials
          KEYCLOAK_ADMIN               = "admin"
          KEYCLOAK_ADMIN_PASSWORD      = "adminpass"

          # Keycloak DB configuration
          KC_DB_URL                    = "jdbc:postgresql://localhost:5432/keycloakdb"
          KC_DB_USERNAME               = "postgres"
          KC_DB_PASSWORD               = "newpassword"
          KC_PROXY                     = "edge"
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
          "traefik.http.routers.keycloak-http.priority=10"
        ]
        provider = "nomad"
      }
    }
  }
}
