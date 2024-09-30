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
        KC_HEALTH_ENABLED        = "true"
        KC_HTTP_ENABLED          = "true"
        KC_PROXY                 = "edge"
        KC_HOSTNAME_STRICT_HTTPS = "false"
        KC_HOSTNAME = "localhost"
        KC_LOG_LEVEL             = "WARN,io.quarkus:INFO,org.infinispan.CONTAINER:INFO"
        KC_DB                    = "postgres"
        KC_DB_URL_HOST           = "localhost"
        KC_DB_URL_PORT           = "5432"
        KC_DB_URL_DATABASE       = "jdbc:postgresql://localhost:5432/model_bazaar"
        KC_DB_USERNAME           = "postgres"
        KC_DB_PASSWORD           = "newpassword"
        KEYCLOAK_ADMIN           = "kc_admin"
        KEYCLOAK_ADMIN_PASSWORD  = "password"
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
