"""
Things we may want to test:

* what do our query times look like with many files vs just a few files
* stress test the cache and generation stuff separately


have a utils to create a deployment, login, etc
"""

import argparse
import json
import random
import sys
from dataclasses import dataclass
from urllib.parse import urljoin

import pandas as pd
import os
import requests
from locust import HttpUser, TaskSet, between, task
from locust.main import main as locust_main
from requests.auth import HTTPBasicAuth


def parse_args():
    parser = argparse.ArgumentParser(add_help=False)
    # parser.add_argument("--host", type=str, default="http://98.82.179.129")
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--deployment_id", type=str, default="7be21289-d3d0-4308-a99f-b14ee07afe90")
    parser.add_argument("--email", type=str, default="david@thirdai.com")
    parser.add_argument("--password", type=str, default="password")
    parser.add_argument(
        "--min_wait",
        type=int,
        default=10,
        help="Minimum wait time between tasks in seconds",
    )
    parser.add_argument(
        "--max_wait",
        type=int,
        default=20,
        help="Maximum wait time between tasks in seconds",
    )

    args, unknown = parser.parse_known_args()

    # Remove our custom args from sys.argv
    sys.argv = [sys.argv[0]] + unknown

    return args


args = parse_args()


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


login_details = Login.with_email(
    base_url=urljoin(args.host, "/api/"),
    email=args.email,
    password=args.password,
)
auth_token = login_details.access_token


class ModelBazaarLoadTest(TaskSet):
    # @task(1)
    # def test_predict(self):
    #     query = "test query"
    #     headers = {
    #         "Authorization": f"Bearer {auth_token}",
    #     }

    #     response = self.client.post(
    #         f"/{args.deployment_id}/search",
    #         json={"query": query, "top_k": 5},
    #         headers=headers,
    #         timeout=60,
    #     )

    @task(1)
    def test_chat(self):
        query = "who is eligible"
        # query = "who is eligible, please be as verbose as possible"
        headers = {
            "Authorization": f"Bearer {auth_token}",
        }

        response = self.client.post(
            f"/{args.deployment_id}/search",
            json={"query": query, "top_k": 5},
            headers=headers,
            timeout=60,
        )

        references = [{"text": x["text"], "source": x["source"]} for x in response.json()["data"]["references"]]

        import time
        start = time.time()
        response = self.client.post(
            f"/llm-dispatch/generate",
            json={"query": query, "references": references, "key": "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"},
            headers=headers,
            timeout=60,
        )
        print(time.time() - start)


    # @task(1)
    # def test_insert(self):
    #     def doc_dir():
    #         return os.path.join(
    #             os.path.dirname(os.path.dirname(__file__)),
    #             "thirdai_platform/train_job/sample_docs",
    #         )
        
    #     documents = [
    #         {"path": "mutual_nda.pdf", "location": "local"},
    #         {"path": "four_english_words.docx", "location": "local"},
    #         {"path": "supervised.csv", "location": "local"},
    #     ]

    #     files = [
    #         *[
    #             ("files", open(os.path.join(doc_dir(), doc["path"]), "rb"))
    #             for doc in documents
    #         ],
    #         ("documents", (None, json.dumps({"documents": documents}), "application/json")),
    #     ]

    #     res = self.client.post(
    #         f"/{args.deployment_id}/insert",
    #         files=files,
    #     )


class WebsiteUser(HttpUser):
    tasks = [ModelBazaarLoadTest]
    wait_time = between(args.min_wait, args.max_wait)
    host = args.host


if __name__ == "__main__":
    locust_main()



# python version cant be 3.8

# pip3 install locust --upgrade --no-cache-dir --force-reinstall

# locust -f stress_test_deployment.py --master

# for i in {1..45}; do
#   locust -f - --worker --master-host=192.168.1.6 &
# done