from airflow.models import BaseOperator
from airflow.utils.decorators import apply_defaults
import requests

class NomadOperator(BaseOperator):
    @apply_defaults
    def __init__(self, job_spec, nomad_endpoint="http://localhost:4646", *args, **kwargs):
        super(NomadOperator, self).__init__(*args, **kwargs)
        self.job_spec = job_spec
        self.nomad_endpoint = nomad_endpoint

    def execute(self, context):
        submit_url = f"{self.nomad_endpoint}/v1/jobs"
        headers = {"Content-Type": "application/json"}

        # Wrap the job spec with the "Job" key
        payload = {"Job": self.job_spec}

        response = requests.post(submit_url, headers=headers, json=payload)
        response.raise_for_status()
        self.log.info("Job submitted to Nomad successfully.")
