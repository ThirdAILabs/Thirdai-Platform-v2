job "keycloak" {
  datacenters = ["dc1"]

  group "keycloak-group" {
    count = 1

    network {
      port "keycloak-http" {
        to = 8180
      }
    }

    task "keycloak" {
      driver = "docker"

      config {
        image = "quay.io/keycloak/keycloak:22.0"
        ports = ["keycloak-http"]
        args  = ["start", "--debug", "--http-port=8180"]
      }

      env {
        KC_HEALTH_ENABLED            = "true"
        KC_METRICS_ENABLED           = "true"
        KC_HTTP_ENABLED              = "true"
        KC_HOSTNAME_STRICT_HTTPS      = "false"
        KEYCLOAK_SSL_REQUIRED        = "none"
        KC_HOSTNAME_STRICT_BACKCHANNEL = "false"
        KC_HOSTNAME_URL            = "http://localhost/keycloak"
        KC_HOSTNAME_ADMIN_URL             = "http://localhost/keycloak"
        KC_HTTP_RELATIVE_PATH     =  "/keycloak"
        KC_PROXY                 = "edge"
        KC_HOSTNAME_STRICT      =false

        # Database connection
        DB_VENDOR                    = "postgres"
        DB_ADDR                      = "172.17.0.1"
        DB_DATABASE                  = "keycloakdb"
        DB_USER                      = "postgres"
        DB_PASSWORD                  = "newpassword"

        KEYCLOAK_ADMIN               = "temp_admin"
        KEYCLOAK_ADMIN_PASSWORD      = "password"
      }

      resources {
        cpu    = 500
        memory = 2048
      }

      service {
        name = "keycloak"
        port = "keycloak-http"
        provider = "nomad"
        tags = [
          "traefik.enable=true",
          "traefik.http.routers.keycloak-http.rule=PathPrefix(`/keycloak`)",
          "traefik.http.routers.keycloak-http.priority=10"
        ]
      }
    }
  }
}