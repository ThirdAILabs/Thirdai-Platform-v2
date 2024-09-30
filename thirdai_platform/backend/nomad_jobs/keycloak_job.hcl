job "keycloak" {
  datacenters = ["dc1"]  # Specify the datacenter(s)

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
          KC_HOSTNAME                  = "localhost"
          KC_HOSTNAME_PORT             = "8080"

          # Database connection
          DB_VENDOR                    = "postgres"
          DB_ADDR                      = "localhost"  # Change to localhost since you're running Postgres locally
          DB_DATABASE                  = "keycloakdb"
          DB_USER                      = "postgres"
          DB_PASSWORD                  = "newpassword"

          # Keycloak Admin credentials
          KEYCLOAK_ADMIN               = "admin"
          KEYCLOAK_ADMIN_PASSWORD      = "adminpass"

          # Regular Keycloak user credentials
          # TODO(pratik): See whether we need them or not.
          KEYCLOAK_USER                = "user"
          KEYCLOAK_PASSWORD            = "userpass"

          # Keycloak DB configuration (PostgreSQL)
          KC_DB                        = "postgres"
          KC_DB_URL                    = "jdbc:postgresql://localhost:5432/keycloakdb"  # Change postgres to localhost here
          KC_DB_USERNAME               = "postgres"
          KC_DB_PASSWORD               = "newpassword"
      }



      resources {
        cpu    = 500
        memory = 512
      }

      # Service registration for Traefik
      service {
        name = "keycloak"
        port = "keycloak-http"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.keycloak-http.rule=PathPrefix(`/keycloak/`)",
          "traefik.http.routers.keycloak-http.priority=10"
        ]
        provider = "nomad"  # Ensure that Nomad, not Consul, is used for service discovery
      }
    }

    service {
      provider = "nomad"
    }
  }

}
