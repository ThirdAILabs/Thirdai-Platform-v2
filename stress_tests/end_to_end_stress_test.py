import argparse
import os
import subprocess
from urllib.parse import urljoin

from client.bazaar import ModelBazaar


class StressTestConfig:
    name: str
    s3_url: str


class SinglePDFConfig(StressTestConfig):
    name: str = "single-pdf"
    s3_url: str = (
        "/home/david/ThirdAI-Platform/thirdai_platform/train_job/sample_docs/mutual_nda.pdf"
    )


class LargeCSVConfig(StressTestConfig):
    name: str = "large-csv"
    s3_url: str = ""


class ManyFilesConfig(StressTestConfig):
    name: str = "many-files"
    s3_url: str = ""


configs = {
    config.name: config
    for config in [SinglePDFConfig(), LargeCSVConfig(), ManyFilesConfig()]
}


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str)
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--email", type=str, default="david@thirdai.com")
    parser.add_argument("--username", type=str, default="david")
    parser.add_argument("--password", type=str, default="password")
    parser.add_argument("--autoscaling_enabled", type=bool, default=False)
    parser.add_argument("--users", type=int, default=100)
    parser.add_argument("--spawn_rate", type=int, default=10)
    parser.add_argument("--run_time", type=int, default=60)  # in seconds
    args = parser.parse_args()

    return args


def create_deployment(client, config, autoscaling_enabled):
    client.log_in(args.email, args.password)
    model_object = client.train(
        f"stress_test_{config.name}",
        unsupervised_docs=[config.s3_url],
        model_options={"ndb_options": {"ndb_sub_type": "v2"}},
        supervised_docs=[],
        doc_type="local",
    )
    model_identifier = model_object.model_identifier
    ndb_client = client.deploy(
        model_identifier, autoscaling_enabled=autoscaling_enabled
    )
    return model_identifier, ndb_client


def run_stress_test(args, deployment_id):
    print("Running Stress Test\n")
    folder = os.path.dirname(__file__)
    script_path = os.path.join(folder, "stress_test_deployment.py")
    result = subprocess.run(
        f"locust -f {script_path} --headless --users {args.users} --spawn-rate {args.spawn_rate} --run-time {args.run_time} --host {args.host} --deployment_id {deployment_id} --email {args.email} --password {args.password}",
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        shell=True,
    )

    if result.returncode != 0:
        print("Error occurred:", result.stderr)
    else:
        print(result.stdout)


def check_nomad_job_status(model_id):
    # check traefik, model bazaar, and deployment jobs and see if they are down
    pass


def main(args):
    config = configs[args.config]

    client = ModelBazaar(urljoin(args.host, "/api/"))
    client.log_in(args.email, args.password)

    model_name = f"stress_test_{config.name}"
    model_identifier = f"{args.username}/{model_name}"

    errors = []
    ndb_client = None
    try:
        model_object = client.train(
            model_name,
            unsupervised_docs=[config.s3_url],
            model_options={"ndb_options": {"ndb_sub_type": "v2"}},
            supervised_docs=[],
            doc_type="local",
        )

        ndb_client = client.deploy(
            model_identifier, autoscaling_enabled=args.autoscaling_enabled
        )

        run_stress_test(args, ndb_client.model_id)

        check_nomad_job_status(model_object.model_id)
    except Exception as e:
        errors.append(f"Testing error: {e}")
        raise
    finally:
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

        if errors:
            raise RuntimeError(
                f"Raised {len(errors)} errors: \n - " + "\n - ".join(errors)
            )


if __name__ == "__main__":
    args = parse_args()

    main(args)
