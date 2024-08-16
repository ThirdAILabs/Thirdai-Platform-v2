import json
import os
from dataclasses import dataclass
from pathlib import Path
from typing import List, Optional, Union
from urllib.parse import urljoin

from pydantic import BaseModel, ValidationError
from requests.auth import HTTPBasicAuth

from .utils import (
    create_model_identifier,
    http_delete_with_error,
    http_get_with_error,
    http_post_with_error,
)


class BazaarEntry(BaseModel):
    name: str
    author_username: str
    identifier: str
    trained_on: Optional[str] = None
    num_params: int
    size: int
    size_in_memory: int
    hash: str
    domain: str
    description: Optional[str] = None
    is_indexed: bool = False
    publish_date: str
    author_email: str
    access_level: str = "public"
    thirdai_version: str

    @staticmethod
    def from_dict(entry):
        return BazaarEntry(
            name=entry["model_name"],
            author_username=entry["username"],
            identifier=create_model_identifier(
                model_name=entry["model_name"], author_username=entry["username"]
            ),
            trained_on=entry["trained_on"],
            num_params=entry["num_params"],
            size=entry["size"],
            size_in_memory=entry["size_in_memory"],
            hash=entry["hash"],
            domain=entry["domain"],
            description=entry["description"],
            is_indexed=entry["is_indexed"],
            publish_date=entry["publish_date"],
            author_email=entry["user_email"],
            access_level=entry["access_level"],
            thirdai_version=entry["thirdai_version"],
        )

    @staticmethod
    def bazaar_entry_from_json(json_entry):
        try:
            loaded_entry = BazaarEntry.from_dict(json_entry)
            return loaded_entry
        except ValidationError as e:
            print(f"Validation error: {e}")
            return None


@dataclass
class Login:
    base_url: str
    username: str
    access_token: str

    @staticmethod
    def with_email(
        base_url: str,
        email: str,
        password: str,
    ):
        # We are using HTTPBasic Auth in backend. update this when we change the Authentication in Backend.
        response = http_get_with_error(
            urljoin(base_url, "user/email-login"),
            auth=HTTPBasicAuth(email, password),
        )

        content = json.loads(response.content)
        username = content["data"]["user"]["username"]
        access_token = content["data"]["access_token"]
        return Login(base_url, username, access_token)


def auth_header(access_token):
    return {
        "Authorization": f"Bearer {access_token}",
    }


def relative_path_depth(child_path: Path, parent_path: Path):
    child_path, parent_path = child_path.resolve(), parent_path.resolve()
    relpath = os.path.relpath(child_path, parent_path)
    if relpath == ".":
        return 0
    else:
        return 1 + relpath.count(os.sep)


# Use this decorator for any function to enforce users use only after login.
def login_required(func):
    def wrapper(self, *args, **kwargs):
        if not self.is_logged_in():
            raise PermissionError(
                "This method requires login, please use '.login()' first then try again."
            )
        return func(self, *args, **kwargs)

    return wrapper


class Bazaar:
    def __init__(
        self,
        base_url,
        cache_dir: Union[Path, str],
    ):
        cache_dir = Path(cache_dir)
        if not os.path.exists(cache_dir):
            os.makedirs(cache_dir)
        self._cache_dir = cache_dir
        if not base_url.endswith("/api/"):
            raise ValueError("base_url must end with '/api/'.")
        self._base_url = base_url
        self._login_instance = None

    def signup(self, email, password, username):
        json_data = {
            "username": username,
            "email": email,
            "password": password,
        }

        response = http_post_with_error(
            urljoin(self._base_url, "user/email-signup-basic"),
            json=json_data,
        )

        print(
            f"Successfully signed up. Please check your email ({email}) to verify your account."
        )

    def login(self, email, password):
        self._login_instance = Login.with_email(self._base_url, email, password)

    def add_global_admin(self, email):
        response = http_post_with_error(
            urljoin(self._base_url, "user/add-global-admin"),
            json={"email": email},
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def delete_user(self, email):
        response = http_delete_with_error(
            urljoin(self._base_url, "user/delete-user"),
            json={"email": email},
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def add_secret_key(self, key, value):
        secret_data = {"key": key, "value": value}

        response = http_post_with_error(
            urljoin(self._base_url, "vault/add-secret"),
            json=secret_data,
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def get_secret_key(self, key):
        secret_data = {"key": key}

        response = http_get_with_error(
            urljoin(self._base_url, "vault/get-secret"),
            json=secret_data,
            headers=auth_header(self._login_instance.access_token),
        )

        return response

    def create_team(self, name):
        response = http_post_with_error(
            urljoin(self._base_url, "team/create-team"),
            params={"name": name},
            headers=auth_header(self._login_instance.access_token),
        )
        response_content = json.loads(response.content)
        return response_content["data"]["team_id"]

    def remove_user_from_team(self, user_email, team_id):
        response = http_post_with_error(
            urljoin(self._base_url, "team/remove-user-from-team"),
            params={"email": user_email, "team_id": team_id},
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def add_user_to_team(self, user_email, team_id):
        response = http_post_with_error(
            urljoin(self._base_url, "team/add-user-to-team"),
            params={"email": user_email, "team_id": team_id},
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def assign_team_admin(self, user_email, team_id):
        response = http_post_with_error(
            urljoin(self._base_url, "team/assign-team-admin"),
            params={"email": user_email, "team_id": team_id},
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def delete_team(self, team_id):
        response = http_delete_with_error(
            urljoin(self._base_url, "team/delete-team"),
            params={"team_id": team_id},
            headers=auth_header(self._login_instance.access_token),
        )
        return response

    def is_logged_in(self):
        return self._login_instance is not None
