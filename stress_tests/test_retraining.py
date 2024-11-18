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
    parser.add_argument("--docs_folder", type=str)
    parser.add_argument("--docs_per_insertion", type=int, default=3)
    args = parser.parse_args()
    return args

args = parse_args()


client = ModelBazaar(urljoin(args.host, "/api/"))
client.log_in(args.email, args.password)

config = configs[args.config]
model_name = f"stress_test_{config.name}_{uuid.uuid4()}"
model_identifier = f"{client._username}/{model_name}"

doc_dir = "/home/david/intuit_csvs"
unsupervised_docs = [
    os.path.join(doc_dir, f) for f in os.listdir(doc_dir) if os.path.isfile(os.path.join(doc_dir, f))
]

model_object = client.train(
    model_name,
    unsupervised_docs=unsupervised_docs,
    model_options={"ndb_options": {"ndb_sub_type": "v2"}},
    supervised_docs=[],
    doc_type="local",
)

ndb_client = client.deploy(
    model_identifier, autoscaling_enabled=True
)

if args.docs_folder is not None:
    doc_dir = args.docs_folder
else:
    doc_dir = os.path.join(
        os.path.dirname(os.path.abspath(__file__)),
        "../thirdai_platform/train_job/sample_docs/",
    )

files = [
    f for f in os.listdir(doc_dir) if os.path.isfile(os.path.join(doc_dir, f))
]

possible_docs = [
    {"path": os.path.join(doc_dir, filename), "location": "local"}
    for filename in files
]

documents = random.sample(possible_docs, args.docs_per_insertion)

ndb_client.insert(documents=documents)


start = time.time()
client.retrain_ndb(new_model_name=f"new_model_{uuid.uuid4()}", base_model_identifier=model_identifier)
end = time.time()
print("Retraining Time", end - start)


client.undeploy(ndb_client)
client.delete(model_identifier)