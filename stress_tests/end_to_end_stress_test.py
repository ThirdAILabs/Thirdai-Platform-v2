import argparse
import logging
import os
import shutil
import subprocess
import zipfile
from typing import List
from urllib.parse import urljoin

import boto3
import requests
from botocore.client import Config

from client.bazaar import ModelBazaar

logging.basicConfig(level=logging.INFO)


class StressTestConfig:
    name: str
    docs_s3_uris: List[str]
    queries_s3_uri: str


class SinglePDFConfig(StressTestConfig):
    name: str = "small-pdf"
    docs_s3_uris: List[str] = [
        "s3://thirdai-datasets/ThirdAI-Platform-Stress-Testing/small-pdf/DARPA-SN-24-118.pdf"
    ]
    queries_s3_uri: str = (
        "s3://thirdai-datasets/ThirdAI-Platform-Stress-Testing/small-pdf/queries.csv"
    )


class LargeCSVConfig(StressTestConfig):
    name: str = "large-csv"
    docs_s3_uris: List[str] = [
        "s3://thirdai-datasets/ThirdAI-Platform-Stress-Testing/large-csv/pubmed_1M.csv"
    ]
    queries_s3_uri: str = (
        "s3://thirdai-datasets/ThirdAI-Platform-Stress-Testing/large-csv/queries.csv"
    )


class ManyFilesConfig(StressTestConfig):
    name: str = "many-files"
    docs_s3_uris: List[str] = [
        "s3://thirdai-datasets/ThirdAI-Platform-Stress-Testing/many-files/docs",
        "s3://novatris-demo/all_icml_files",
    ]
    queries_s3_uri: str = (
        "s3://thirdai-datasets/ThirdAI-Platform-Stress-Testing/many-files/queries.csv"
    )


configs = {
    config.name: config
    for config in [SinglePDFConfig(), LargeCSVConfig(), ManyFilesConfig()]
}


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str)
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--email", type=str, required=True)
    parser.add_argument("--password", type=str, required=True)
    parser.add_argument("--autoscaling_enabled", action="store_true")
    parser.add_argument("--users", type=int, default=1000)
    parser.add_argument("--spawn_rate", type=int, default=30)
    parser.add_argument("--run_time", type=int, default=60)  # in seconds
    parser.add_argument("--no_cleanup", action="store_true")
    args = parser.parse_args()

    return args


def get_nomad_job(job_id, nomad_endpoint):
    headers = {"X-Nomad-Token": os.getenv("TASK_RUNNER_TOKEN")}
    response = requests.get(
        urljoin(nomad_endpoint, f"v1/job/{job_id}"), headers=headers
    )
    if response.status_code == 200:
        return response.json()
    return None


def download_query_file(queries_s3_uri):
    aws_access_key = os.getenv("AWS_ACCESS_KEY", None)
    aws_secret_access_key = os.getenv("AWS_ACCESS_SECRET", None)
    region_name = os.getenv("AWS_REGION_NAME", None)
    s3_client = boto3.client(
        "s3",
        aws_access_key_id=aws_access_key,
        aws_secret_access_key=aws_secret_access_key,
        config=Config(
            retries={"max_attempts": 10, "mode": "standard"},
            connect_timeout=5,
            read_timeout=60,
            signature_version="s3v4",
        ),
        region_name=region_name,
    )
    folder = os.path.dirname(__file__)
    query_file_dest_path = os.path.join(folder, "queries.csv")
    bucket_name = queries_s3_uri.strip("s3://").split("/")[0]
    source_path = "/".join(queries_s3_uri.strip("s3://").split("/")[1:])
    s3_client.download_file(bucket_name, source_path, query_file_dest_path)
    return query_file_dest_path


def run_stress_test(args, query_file, deployment_id):
    folder = os.path.dirname(__file__)
    script_path = os.path.join(folder, "stress_test_deployment.py")
    command = (
        f"locust -f {script_path} --headless --users {args.users} --spawn-rate {args.spawn_rate} --run-time {args.run_time} --host {args.host} --deployment_id {deployment_id} --email {args.email} --password {args.password} --query_file {query_file}",
    )
    result = subprocess.run(command, shell=True)
    status = "success" if result.returncode == 0 else "failed"
    return status


# check traefik, model bazaar, and deployment jobs and see if they are down
def check_nomad_job_status(deployment_id):
    os.environ["NOMAD_ENDPOINT"] = "http://localhost:4646/"

    data = get_nomad_job("traefik", os.getenv("NOMAD_ENDPOINT"))
    if data is not None and data["Status"] == "dead":
        raise ValueError(f"traefik job has died.")

    data = get_nomad_job("model_bazaar", os.getenv("NOMAD_ENDPOINT"))
    if data is not None and data["Status"] == "dead":
        raise ValueError(f"model_bazaar job has died.")

    data = get_nomad_job(f"deployment-{deployment_id}", os.getenv("NOMAD_ENDPOINT"))
    if data is None:
        raise ValueError(f"deployment job not found.")
    if data["Status"] == "dead":
        raise ValueError(f"deployment job has died.")


def main(args):
    config = configs[args.config]

    client = ModelBazaar(urljoin(args.host, "/api/"))
    client.log_in(args.email, args.password)

    model_name = f"stress_test_{config.name}"
    model_identifier = f"{client._username}/{model_name}"

    query_file = download_query_file(config.queries_s3_uri)

    errors = []
    ndb_client = None
    try:
        model_object = client.train(
            model_name,
            unsupervised_docs=config.docs_s3_uris,
            model_options={"ndb_options": {"ndb_sub_type": "v2"}},
            supervised_docs=[],
            doc_type="s3",
        )

        ndb_client = client.deploy(
            model_identifier, autoscaling_enabled=args.autoscaling_enabled
        )

        logging.info("Running Stress Test\n\n\n")
        stress_test_status = run_stress_test(args, query_file, ndb_client.model_id)
        logging.info(f"\nStress test status: {stress_test_status}\n\n\n")

        logging.info("Printing Deployment Logs:\n\n")
        log_file_path = client.logs(model_object)
        with zipfile.ZipFile(log_file_path, "r") as z:
            z.extractall("logs")
        with open("logs/deployment.log", "r") as f:
            lines = f.read_lines()
            for line in lines:
                if not line.strip().endswith("200"):
                    logging.info(line.strip())

        check_nomad_job_status(ndb_client.model_id)

    except Exception as e:
        errors.append(f"Testing error: {e}")
        raise
    finally:
        if not args.no_cleanup:
            if ndb_client:
                try:
                    client.undeploy(ndb_client)
                except Exception as e:
                    errors.append(f"Undeploy error: {e}")

            # This gives a permissions denied error when run locally without running the
            # backend in sudo. TODO fix this issue
            try:
                client.delete(model_identifier)
            except Exception as e:
                errors.append(f"Delete error: {e}")

            os.remove(query_file)
            shutil.rmtree("logs")
            shutil.rmtree("bazaar_cache")

            if errors:
                raise RuntimeError(
                    f"Raised {len(errors)} errors: \n - " + "\n - ".join(errors)
                )


if __name__ == "__main__":
    args = parse_args()

    main(args)
