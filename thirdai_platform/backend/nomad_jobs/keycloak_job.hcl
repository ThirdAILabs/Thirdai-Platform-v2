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
        image = "quay.io/keycloak/keycloak:26.0.0"
        ports = ["keycloak-http"]
        args  = ["start", "--debug", "--http-port=8180"]
      }

      env {
        KC_HTTP_ENABLED                   = "true"
        KC_HOSTNAME_STRICT_BACKCHANNEL    = "false"
        KC_HOSTNAME                       = "http://localhost/keycloak"
        KC_HOSTNAME_ADMIN                 = "http://localhost/keycloak"
        KC_HOSTNAME_BACKCHANNEL_DYNAMIC   = "true"
        KC_HTTP_RELATIVE_PATH             = "/keycloak"
        KC_HOSTNAME_STRICT                = "true"

        # Database connection
        DB_VENDOR                    = "postgres"
        DB_ADDR                      = "172.17.0.1"
        DB_DATABASE                  = "keycloakdb"
        DB_USER                      = "postgres"
        DB_PASSWORD                  = "newpassword"

        KC_BOOTSTRAP_ADMIN_USERNAME  = "temp_admin"
        KC_BOOTSTRAP_ADMIN_PASSWORD  = "password"
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