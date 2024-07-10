import os
from airflow import DAG
from airflow.utils.dates import days_ago
from datetime import timedelta
from nomad_operator import NomadOperator

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

# There is some urllib package issue when we try sending a post request it
# throws  "Negsignal.SIGSEGV".  The proposed solution from stackoverflow
# https://stackoverflow.com/questions/73582293/airflow-external-api-call-gives-negsignal-sigsegv-error
os.environ["no_proxy"] = "*"

with DAG(
    'nomad_example_dag',
    default_args=default_args,
    description='A simple Nomad DAG',
    tags=['example'],
) as dag:
    job_spec = {
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

    nomad_task = NomadOperator(
        task_id='submit_nomad_job',
        job_spec=job_spec,
    )
