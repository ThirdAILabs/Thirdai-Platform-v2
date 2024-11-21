import time
import datetime
import random
import argparse
import logging
import os
import shutil
import subprocess
import zipfile
from typing import List
from urllib.parse import urljoin
import uuid

import boto3
import requests
from botocore.client import Config

from client.bazaar import ModelBazaar


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--email", type=str, default="david@thirdai.com")
    parser.add_argument("--password", type=str, default="password")
    parser.add_argument("--deployment_id", type=str)
    parser.add_argument("--doc_size_mb", type=int, choices=[5, 100, 500], default=1)
    args = parser.parse_args()
    return args

args = parse_args()


if not args.deployment_id:
    client = ModelBazaar(urljoin(args.host, "/api/"))
    client.log_in(args.email, args.password)

    model_name = f"stress_test_{uuid.uuid4()}"
    model_identifier = f"{client._username}/{model_name}"

    doc_dir = "/home/david/intuit_csvs"

    if args.doc_size_mb == 5:
        files = [os.path.join(doc_dir, "2024_KYI_PER.csv")]
    elif args.doc_size_mb == 100:
        files = [
            os.path.join(doc_dir, f) for f in os.listdir(doc_dir) if os.path.isfile(os.path.join(doc_dir, f))
        ]
    elif args.doc_size_mb == 500:
        doc_dir = "/home/david/realistic_intuit_dataset"
        files = [
            os.path.join(doc_dir, f)  for f in os.listdir(doc_dir) if os.path.isfile(os.path.join(doc_dir, f))
        ]


    model_object = client.train(
        model_name,
        unsupervised_docs=files,
        model_options={"ndb_options": {"ndb_sub_type": "v2"}},
        supervised_docs=[],
        doc_type="local",
    )

    ndb_client = client.deploy(
        model_identifier, autoscaling_enabled=True
    )

    time.sleep(60)

    deployment_id = ndb_client.model_id
else: 
    deployment_id = args.deployment_id

# exit()

folder = os.path.dirname(__file__)
script_path = os.path.join(folder, "stress_test_deployment.py")


def run_stress_test_with_qps(qps, constrained=False):
    if constrained: 
        predict_weight = 0
        constrained_weight = 1
    else:
        predict_weight = 1
        constrained_weight = 0

    command = (
        f"locust -f {script_path} --headless --users {qps} --spawn-rate {5} --run-time 180 --host {args.host} --deployment_id {deployment_id} --email {args.email} --password {args.password} --min_wait 0.8 --max_wait 1.2 --predict_weight {predict_weight} --constrained_search_weight {constrained_weight}",
    )

    result = subprocess.run(
        command,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        shell=True,
    )

    output_lines = result.stdout.splitlines()[-50:]
    print(f"QPS: {qps} constrained: {constrained}")
    for line in output_lines:
        print(line)
    print("\n\n\n")
    time.sleep(3)

run_stress_test_with_qps(10, constrained=False)
run_stress_test_with_qps(50, constrained=False)
run_stress_test_with_qps(100, constrained=False)

run_stress_test_with_qps(10, constrained=True)
run_stress_test_with_qps(50, constrained=True)
run_stress_test_with_qps(100, constrained=True)

# de907d54-833f-45a8-afcc-4da33ddfc681 1 MB
# 8a77f33e-9dcf-49ed-896e-27bf1b8be6e6 100 MB
# 7bd852da-dec5-476d-ae0b-b8424034f59d 500 MB