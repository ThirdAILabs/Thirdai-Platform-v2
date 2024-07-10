from datetime import datetime, timedelta
import os
from airflow import DAG
from airflow.operators.bash import BashOperator
from airflow.operators.python_operator import PythonOperator
from nomad_operator import NomadOperator

# Default arguments for the DAG
default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'start_date': datetime(2024, 7, 10),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

# There is some urllib package issue when we try sending a post request it
# throws  "Negsignal.SIGSEGV".  The proposed solution from stackoverflow
# https://stackoverflow.com/questions/73582293/airflow-external-api-call-gives-negsignal-sigsegv-error
os.environ["no_proxy"] = "*"

# Instantiate the DAG object
dag = DAG(
    'temp',
    default_args=default_args,
    description='A complex DAG with multiple steps',
    schedule_interval=timedelta(days=1),
)

# Task 1: BashOperator to print a message
task_print_message = BashOperator(
    task_id='print_message',
    bash_command='echo "Hello from Airflow!"',
    dag=dag,
)

script_path = os.path.join(os.path.dirname(__file__), "run.py")
requiremnts_path = os.path.join(os.path.dirname(__file__), "requirements.txt")

# Task 3: NomadOperator to run a job on Nomad (executing the Python script)
nomad_job = {
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
                    "Config" : {
                        "command" : "/bin/bash",
                        "args"     : ["-c", f"pip install -r {requiremnts_path} && python {script_path}"],
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

task_run_nomad_job = NomadOperator(
    task_id='run_nomad_job',
    job_spec=nomad_job,
    dag=dag,
)

# Define the task dependencies
task_print_message >> task_run_nomad_job
