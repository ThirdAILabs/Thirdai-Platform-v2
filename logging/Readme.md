# Setup Instructions to run Grafana and Loki for logs monitoring

1. **Start Nomad:**
   - Open a terminal session and run:
     ```
     sudo nomad agent -config="agent.hcl"
     ```

2. **Launch Nomad Jobs**
   - Launch individual nomad jobs:
     ```
     nomad job run grafana.nomad.hcl
     nomad job run loki.nomad.hcl
     nomad job run traefik.nomad.hcl
     ```

3. **Port forward your localhost port**
    - To run grafana dashboard, run on local machine
    ```
    ssh -L <local_port>:<private_ip>:80 -J <jump_host> username@<private_ip>
    ```