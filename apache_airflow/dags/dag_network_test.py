import os
from airflow import DAG
from airflow.operators.python_operator import PythonOperator
import requests

# There is some urllib package issue when we try sending a post request it
# throws  "Negsignal.SIGSEGV".  The proposed solution from stackoverflow
# https://stackoverflow.com/questions/73582293/airflow-external-api-call-gives-negsignal-sigsegv-error
os.environ["no_proxy"] = "*"

def test_network_call():
    try:
        response = requests.get("https://httpbin.org/get")
        response.raise_for_status()
        print(f"Network call successful: {response.json()}")
    except Exception as e:
        print(f"Network call failed: {e}")
        raise

with DAG(
    'network_test_dag',
) as dag:
    network_test_task = PythonOperator(
        task_id='test_network_call',
        python_callable=test_network_call,
    )

    network_test_task
