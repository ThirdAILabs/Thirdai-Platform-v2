.PHONY: monitors

monitors:
	cd job_files && nomad run victoriametrics.hcl
	cd job_files && nomad run vmagnet.hcl
	cd job_files && nomad run grafana.hcl
