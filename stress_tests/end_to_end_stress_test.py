import argparse
import os
import subprocess
from typing import List


class StressTestConfig:
    config_name: str
    files: List[str]


class SinglePDFConfig(StressTestConfig):
    config_name: str = "single-pdf"


configs = {}


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str)
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--email", type=str, default="david@thirdai.com")
    parser.add_argument("--password", type=str, default="password")
    parser.add_argument("--users", type=int, default=100)
    parser.add_argument("--spawn_rate", type=int, default=10)
    args = parser.parse_args()

    return args


def run_stress_test(args, deployment_id):
    print("Running Stress Test\n")
    folder = os.path.dirname(__file__)
    script_path = os.path.join(folder, "stress_test_deployment.py")
    result = subprocess.run(
        f"locust -f {script_path} --headless --users {args.users} --spawn-rate {args.spawn_rate} --host {args.host} --deployment_id {deployment_id} --email {args.email} --password {args.password}",
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )

    if result.returncode != 0:
        print("Error occurred:", result.stderr)
    else:
        print(result.stdout)


def main(args):
    config = configs[args.config_name]

    # download the files, create the deployment

    run_stress_test(args, deployment_id)

    # check the deployment after and delete it if necessary


if __name__ == "__main__":
    args = parse_args()

    main(args)
