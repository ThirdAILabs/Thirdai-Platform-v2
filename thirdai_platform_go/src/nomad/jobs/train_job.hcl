job "{{ .JobName }}" {

  datacenters = ["dc1"]

  type = "batch"

  group "train-job" {
    count = 1

    task "server" {

      {{ if eq .PlatformType "docker" }}
        driver = "docker"
      {{ else if eq .PlatformType "local" }}
        driver = "raw_exec"
      {{ end }}

      env {
        AWS_ACCESS_KEY = "{{ .AwsAccessKey }}"
        AWS_ACCESS_SECRET = "{{ .AwsAccessSecret }}"
      }

      config {
        {{ if eq .PlatformType "docker" }}
          {{ with .Platform }}
          image = "{{ .Registry }}/{{ .ImageName }}:{{ .Tag }}"
          image_pull_timeout = "15m"
          auth {
            username = "{{ .DockerUsername }}"
            password = "{{ .DockerPassword }}"
            server_address = "{{ .Registry }}"
          }
          volumes = [
            "{{ .ShareDir }}:/model_bazaar"
          ]
          {{ end }}
          command = "python3"
          args    = ["-m", "{{ .TrainScript }}", "--config", "{{ .ConfigPath }}"]
        {{ else if eq .PlatformType "local" }}
          command = "/bin/sh"
          args    = ["-c", "cd {{ with .Platform }}{{ .PlatformDir }} && {{ .PythonPath }}{{ end }} -m {{ .TrainScript }} --config {{ .ConfigPath }}"]
        {{ end }}
      }

      resources {
        {{ with .Resources }}
        cpu = {{ .AllocationMhz }}
        memory = {{ .AllocationMemory }}
        memory_max = {{ .AllocationMemoryMax }}
        {{ end }} 
      }
    }
  }
}