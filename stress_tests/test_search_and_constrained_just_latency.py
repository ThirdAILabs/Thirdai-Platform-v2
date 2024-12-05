from tqdm import tqdm 
from requests.auth import HTTPBasicAuth
import json
import time
from dataclasses import dataclass
import random
import argparse
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
    parser.add_argument("--on_disk", action="store_true")
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
        model_options={"ndb_options": {"ndb_sub_type": "v2"}, "on_disk": args.on_disk},
        supervised_docs=[],
        doc_type="local",
        doc_options={doc: {"csv_metadata_columns": {
            "docId": "string",
            "instruction": "string",
            "filePath": "string",
            "createdTime": "string",
            "Tax Year": "string",
            "Formset": "string",
            "Form Id": "string",
            "Form Field Id": "string",
            "LLM Form Description": "string",
            "LLM Field Description": "string",
            "Field Type": "string",
            "Array": "string",
            "Field Name": "string",
            "Min": "string",
            "Field FullName": "string",
            "Source PT Form": "string",
            "Field Description": "string",
            "SNo": "string",
            "Field ID": "string",
            "CID FormID": "string",
            "Table Id": "string",
            "Export": "string",
            "Max": "string",
            "Link Form": "string",
            "Picklist": "string",
            "Link Field ID": "string",
            "CID TableID": "string",
            "CID FieldID": "string",
        }, "csv_strong_columns": [
            "Field FullName",
            "Field Description",
            "LLM Form Description",
            "LLM Field Description",
        ]} for doc in files}
    )

    ndb_client = client.deploy(
        model_identifier, autoscaling_enabled=True
    )

    deployment_id = ndb_client.model_id
else: 
    deployment_id = args.deployment_id



queries = ["what services and employer finance", "for not should BLANK tax finance return tax income services from this a b i"]

def random_query():
    return random.choice(queries)


@dataclass
class Login:
    base_url: str
    username: str
    access_token: str

    @staticmethod
    def with_email(base_url: str, email: str, password: str):
        response = requests.get(
            urljoin(base_url, "user/email-login"),
            auth=HTTPBasicAuth(email, password),
        )
        if response.status_code < 200 or response.status_code >= 300:
            raise Exception(f"Login failed: {response.status_code}, {response.text}")
        content = json.loads(response.content)
        username = content["data"]["user"]["username"]
        access_token = content["data"]["access_token"]
        return Login(base_url, username, access_token)


base_url = urljoin(args.host, "/api/")

login_details = Login.with_email(
    base_url=base_url,
    email=args.email,
    password=args.password,
)

username = login_details.username
auth_header = {"Authorization": f"Bearer {login_details.access_token}"}


def benchmark_queries(num_queries=10, constrained=False):
    total_time = 0
    for _ in tqdm(list(range(num_queries))):
        query = random_query()
        data = {"query": query, "top_k": 5}
        if constrained:
            data["constraints"] = {
                "Tax Year": {"constraint_type": "EqualTo", "value": 2024},
                # "Formset": {"constraint_type": "EqualTo", "value": "FDI"},
                # "Form Id": {"constraint_type": "AnyOf", "values": ["KYTCS", "HNGI", "ZSCA", "WPASSIVE", "something", "other thing", "this", "A KEY", "ANOTHER KEY", "HAHAHA"]}
            }
        start = time.time()
        response = requests.post(
            urljoin(args.host, f"{deployment_id}/search"),
            json=data,
            headers=auth_header,
        )
        if response.status_code != 200:
            raise ValueError("OOpS")
        total_time += time.time() - start
    print(f"Avg Latency for query type constrained = {constrained}: {1000 * (total_time / num_queries)} ms")


benchmark_queries(num_queries=100, constrained=False)
benchmark_queries(num_queries=100, constrained=True)