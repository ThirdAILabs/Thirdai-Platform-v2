service {
  name     = "backend"
  id       = "backend-local"
  address  = "127.0.0.1"
  port     = 8000
  checks = [
    {
      name     = "backend_http_check"
      http     = "http://127.0.0.1:8000"
      interval = "10s"
    }
  ]
}

service {
  name     = "keycloak"
  id       = "keycloak-local"
  address  = "127.0.0.1"
  port     = 8180
  checks = [
    {
      name     = "keycloak_http_check"
      http     = "http://127.0.0.1:8180"
      interval = "10s"
    }
  ]
}

service {
  name     = "frontend"
  id       = "frontend-local"
  address  = "127.0.0.1"
  port     = 3006
  checks = [
    {
      name     = "frontend_http_check"
      http     = "http://127.0.0.1:3006"
      interval = "10s"
    }
  ]
}
