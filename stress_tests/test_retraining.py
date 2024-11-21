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
    parser.add_argument("--email", type=str, required=True)
    parser.add_argument("--password", type=str, required=True)
    parser.add_argument("--docs_folder", type=str)
    parser.add_argument("--docs_per_insertion", type=int, default=3)
    args = parser.parse_args()
    return args

args = parse_args()


client = ModelBazaar(urljoin(args.host, "/api/"))
client.log_in(args.email, args.password)

model_name = f"stress_test_{uuid.uuid4()}"
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

for doc in documents:
    ndb_client.insert(documents=[doc])


start = time.time()
client.retrain_ndb(new_model_name=f"new_model_{uuid.uuid4()}", base_model_identifier=model_identifier)
end = time.time()
print("Retraining Time", end - start)


client.undeploy(ndb_client)
client.delete(model_identifier)