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
        KC_HOSTNAME_URL              = "http://localhost:8180"

        # Database connection
        DB_VENDOR                    = "postgres"
        DB_ADDR                      = "172.17.0.1"
        DB_DATABASE                  = "keycloakdb"
        DB_USER                      = "postgres"
        DB_PASSWORD                  = "newpassword"

        KC_DB                        = "postgres"
        KC_DB_URL                    = "jdbc:postgresql://172.17.0.1:5432/keycloakdb"
        KC_DB_USERNAME               = "postgres"
        KC_DB_PASSWORD               = "newpassword"

        KEYCLOAK_ADMIN               = "kc_admin"
        KEYCLOAK_ADMIN_PASSWORD      = "password"
      }

      resources {
        cpu    = 500
        memory = 512
      }

      service {
        name = "keycloak"
        port = "keycloak-http"
        provider = "nomad"
      }
    }
  }
}
