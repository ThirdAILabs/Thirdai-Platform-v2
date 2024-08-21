job "vault-env-validation" {
  datacenters = ["dc1"]
  type = "batch"

  group "vault-env-validation-group" {
    count = 1

    task "vault-env-validation-task" {
      driver = "docker"

      config {
        image = "alpine:latest"
        command = "sh"
        args = [
          "-c",
          "if [ \"$AWS_ACCESS_TOKEN\" = \"expected_aws_token_value\" ] && [ \"$OPENAI_API_KEY\" = \"expected_openai_api_key_value\" ]; then echo 'Environment variables are correct'; exit 0; else echo 'Environment variables are incorrect'; exit 1; fi"
        ]
      }

      vault {
        policies = ["nomad-job"]
      }

      template {
        data = <<EOT
{% raw %}
{{ with secret "secret/data/AWS_ACCESS_TOKEN" }}
AWS_ACCESS_TOKEN="{{ .Data.data.value }}"
{{ end }}
{{ with secret "secret/data/OPENAI_API_KEY" }}
OPENAI_API_KEY="{{ .Data.data.value }}"
{{ end }}
{% endraw %}
EOT
        destination = "data.env"
        env         = true
      }

      resources {
        cpu    = 100
        memory = 100
      }
    }
  }
}
