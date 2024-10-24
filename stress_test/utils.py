
import argparse
import json
import random
import sys
from dataclasses import dataclass
from urllib.parse import urljoin

import pandas as pd
import requests
from locust import HttpUser, TaskSet, between, task
from locust.main import main as locust_main
from requests.auth import HTTPBasicAuth


@dataclass
class Login:
    base_url: str
    username: str
    access_token: str


def login_with_email(base_url: str, email: str, password: str):
    requests.get()
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