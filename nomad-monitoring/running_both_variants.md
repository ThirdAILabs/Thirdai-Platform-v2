## Initialization

- Change the `path` in the grafana and victoriametrics `host_volume` in the `agent.hcl` to any your local directory

## start the nomad
```bash
make run-nomad
```
### Running as seperate job
```bash
make deploy
```

### Running as a single job
```bash
nomad run combined.hcl
```

NOTE: 
1. If both the variants are being run on the same machine, there could be conflicts of ports. so better to run on different machine itself.
2. Recommended to delete the content of data folder pointing to the host volumes.
```bash
rm -rf data/*/*
```
