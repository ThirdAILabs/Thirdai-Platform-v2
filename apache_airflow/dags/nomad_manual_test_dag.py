import os
from airflow import DAG
from airflow.operators.python_operator import PythonOperator
import requests

# There is some urllib package issue when we try sending a post request it
# throws  "Negsignal.SIGSEGV".  The proposed solution from stackoverflow
# https://stackoverflow.com/questions/73582293/airflow-external-api-call-gives-negsignal-sigsegv-error
os.environ["no_proxy"] = "*"

def submit_nomad_job():
    job_spec = {
        "Job": {
            "ID": "example",
            "Name": "example",
            "Type": "batch",
            "Datacenters": ["dc1"],
            "TaskGroups": [
                {
                    "Name": "example",
                    "Tasks": [
                        {
                            "Name": "example",
                            "Driver": "raw_exec",
                            "Config": {
                                "command": "/bin/echo",
                                "args": ["Hello, Nomad!"]
                            },
                            "Resources": {
                                "CPU": 500,
                                "MemoryMB": 256
                            }
                        }
                    ]
                }
            ]
        }
    }

    submit_url = "http://localhost:4646/v1/jobs"
    headers = {"Content-Type": "application/json"}

    response = requests.post(submit_url, headers=headers, json=job_spec)
    response.raise_for_status()
    print(f"Job submitted to Nomad successfully: {response.json()}")

with DAG(
    'nomad_manual_test_dag',
) as dag:
    manual_test_task = PythonOperator(
        task_id='submit_nomad_job',
        python_callable=submit_nomad_job
    )

    manual_test_task
