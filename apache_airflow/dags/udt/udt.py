from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python_operator import PythonOperator
import subprocess
import shutil
import os
import tempfile

# Function to create virtual environment
def create_virtualenv(venv_dir):
    subprocess.run(['python3', '-m', 'venv', venv_dir], check=True)

# Function to install requirements in virtual environment
def install_requirements(venv_dir, requirements_file):
    activate_cmd = os.path.join(venv_dir, 'bin', 'activate')
    subprocess.check_call(['source', activate_cmd])
    
    subprocess.check_call(['pip', 'install', '-r', requirements_file])
    

# Function to run script_1.py
def run_script_1():
    from udt.run import run  # Import inside the function
    run()

# Function to delete virtual environment
def delete_virtualenv(venv_dir):
    subprocess.check_call(['deactivate'])
    shutil.rmtree(venv_dir)

# Generate a temporary directory for virtual environment
temp_dir = tempfile.TemporaryDirectory()
venv_dir = temp_dir.name

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'start_date': datetime(2024, 7, 10),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

# Instantiate the DAG object
dag = DAG(
    'udt',
    default_args=default_args,
    description='Example DAG to run script_1.py in a virtual environment',
    schedule_interval=timedelta(days=1),
)

# Task: Create virtual environment
task_create_virtualenv = PythonOperator(
    task_id='create_virtualenv',
    python_callable=create_virtualenv,
    op_args=[venv_dir],
    dag=dag,
)

# Task: Install requirements in virtual environment
task_install_requirements = PythonOperator(
    task_id='install_requirements',
    python_callable=install_requirements,
    op_args=[venv_dir, os.path.join(os.path.dirname(__file__), 'requirements.txt')],
    dag=dag,
)

# Task: PythonOperator to run script_1.py
task_run_script_1 = PythonOperator(
    task_id='run_script_1',
    python_callable=run_script_1,
    dag=dag,
)

# Task: Delete virtual environment
task_delete_virtualenv = PythonOperator(
    task_id='delete_virtualenv',
    python_callable=delete_virtualenv,
    op_args=[venv_dir],
    dag=dag,
)

# Define the task dependencies
task_create_virtualenv >> task_install_requirements >> task_run_script_1 >> task_delete_virtualenv
