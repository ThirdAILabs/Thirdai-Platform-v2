import datetime
import logging
from threading import Lock
from typing import Callable, Dict, List, Tuple
from urllib.parse import urljoin

import fastapi
import requests
from fastapi import status

CREDENTIALS_EXCEPTION = fastapi.HTTPException(
    status_code=fastapi.status.HTTP_401_UNAUTHORIZED,
    detail="Invalid access token.",
    # This header indicates what type of authentication would be required to access
    # the resource.
    # https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
    headers={"WWW-Authenticate": "Bearer"},
)


def optional_token_bearer(request: fastapi.Request) -> str:
    """
    Retrieves the token from the Authorization header if it exists, otherwise returns "None".
    """
    # Attempt to retrieve the Authorization header from the request,
    # and return "None" for token if header doesn't exist.
    auth_header = request.headers.get("Authorization")
    if auth_header:
        try:
            # Extract the token type and the token itself from the header.
            token_type, token = auth_header.split(" ", 1)
            if token_type.lower() == "bearer":
                return token
        except ValueError:
            # Return "None" as a dummy value for the token,
            # since we need some string value for the token permission verification.
            return "None"
    return "None"


def now() -> datetime.datetime:
    """
    Returns the current UTC time without microseconds.
    """
    return datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0)


class Permissions:
    model_bazaar_endpoint: str = None
    model_id: str = None
    # entry_expiration_seconds: number of seconds until the permissions for a
    # token needs to be refreshed. We refresh in case a previously invalid
    # token becomes a valid token.
    entry_expiration_min: int = 5
    expirations: List[Tuple[datetime.datetime, str]] = []
    cache: Dict[str, dict] = {}
    cache_lock = Lock()

    @classmethod
    def init(
        cls, model_bazaar_endpoint: str, model_id: str, entry_expiration_min: int = 5
    ):
        cls.model_bazaar_endpoint = model_bazaar_endpoint
        cls.model_id = model_id
        cls.entry_expiration_min = entry_expiration_min

    @classmethod
    def _clear_expired_entries(cls) -> None:
        """
        Clears expired entries from the cache.
        """
        pos = 0
        curr_time = now()
        for expiration, token in cls.expirations:
            if expiration > curr_time:
                break
            try:
                del cls.cache[token]
            except KeyError:
                pass
            pos += 1
        cls.expirations = cls.expirations[pos:]

    @classmethod
    def _deployment_permissions(cls, token: str):
        deployment_permissions_endpoint = urljoin(
            cls.model_bazaar_endpoint,
            f"api/v2/model/{cls.model_id}/permissions",
        )
        response = requests.get(
            deployment_permissions_endpoint,
            headers={"Authorization": "Bearer " + token},
        )
        if response.status_code == status.HTTP_401_UNAUTHORIZED:
            return {
                "read": False,
                "write": False,
                "exp": now() + datetime.timedelta(minutes=5),
                "override": False,
            }
        elif response.status_code != status.HTTP_200_OK:
            logging.info(response.text)
            return {
                "read": False,
                "write": False,
                "exp": now(),
                "override": False,
            }
        permissions = response.json()
        permissions["exp"] = datetime.datetime.fromisoformat(permissions["exp"])
        return permissions

    @classmethod
    def _get_permissions(cls, token: str) -> Tuple[bool, bool, bool, str]:
        """
        Retrieves permissions for a token, updating the cache if necessary.

        Args:
            token (str): The access token.

        Returns:
            Tuple[bool, bool, bool, str]: Read, write, override permissions and username.
        """
        cls._clear_expired_entries()
        curr_time = now()
        if token not in cls.cache:
            permissions = cls._deployment_permissions(token)
            cls.expirations.append(
                (
                    curr_time + datetime.timedelta(minutes=cls.entry_expiration_min),
                    token,
                )
            )
            cls.cache[token] = permissions
            return (
                permissions["read"],
                permissions["write"],
                permissions["owner"],
                permissions["username"],
            )
        if cls.cache[token]["exp"] <= curr_time:
            return False, False, False, "unknown"
        permissions = cls.cache[token]
        return (
            permissions["read"],
            permissions["write"],
            permissions["owner"],
            permissions["username"],
        )

    @classmethod
    def verify_permission(cls, permission_type: str = "read") -> Callable[[str], str]:
        """
        Creates a function that verifies a specific permission type for the token.

        Args:
            permission_type (str): The type of permission to verify (read, write, owner).

        Returns:
            Callable: A function that takes the token and checks the permission.
        """

        def dependency(token: str = fastapi.Depends(optional_token_bearer)) -> str:
            with cls.cache_lock:
                permissions = cls._get_permissions(token)
                permission_map = {
                    "read": permissions[0],
                    "write": permissions[1],
                    "owner": permissions[2],
                }
                if not permission_map.get(permission_type):
                    raise CREDENTIALS_EXCEPTION
                return token

        return dependency

    @classmethod
    def check_permission(cls, token: str, permission_type: str = "read") -> bool:
        """
        Checks if a specific permission type is granted for the token.

        Args:
            token (str): The access token.
            permission_type (str): The type of permission to check (read, write, owner).

        Returns:
            bool: True if the token has the required permission, False otherwise.
        """
        with cls.cache_lock:
            permissions = cls._get_permissions(token)
            permission_map = {
                "read": permissions[0],
                "write": permissions[1],
                "owner": permissions[2],
            }
            return permission_map.get(permission_type, False)
