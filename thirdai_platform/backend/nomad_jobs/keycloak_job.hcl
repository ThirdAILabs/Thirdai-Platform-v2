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
        network_mode = "host"
      }

      env {
          KC_HEALTH_ENABLED            = "true"
          KC_METRICS_ENABLED           = "true"
          KC_HTTP_ENABLED              = "true"
          KC_HOSTNAME_STRICT_HTTPS      = "false"
          KEYCLOAK_SSL_REQUIRED        = "none"
          KC_HOSTNAME_STRICT_BACKCHANNEL = "false"
          KC_HOSTNAME                  = "127.0.0.1"
          KC_HOSTNAME_PORT             = "8180"

          # Database connection
          DB_VENDOR                    = "postgres"
          DB_ADDR                      = "127.0.0.1"  # No change needed, 127.0.0.1 should work in host network
          DB_DATABASE                  = "keycloakdb"
          DB_USER                      = "postgres"
          DB_PASSWORD                  = "newpassword"

          # Keycloak Admin credentials
          KEYCLOAK_ADMIN               = "admin"
          KEYCLOAK_ADMIN_PASSWORD      = "adminpass"

          # Keycloak DB configuration
          KC_DB                        = "postgres"
          KC_DB_URL                    = "jdbc:postgresql://127.0.0.1:5432/keycloakdb"
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
          "traefik.http.routers.keycloak-http.priority=10"
        ]
        provider = "nomad"
      }
    }
  }
}
