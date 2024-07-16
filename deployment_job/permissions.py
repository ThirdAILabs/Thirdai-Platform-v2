import datetime
from threading import Lock
from typing import Dict, List, Tuple

import fastapi
from utils import now
from variables import GeneralVariables

CREDENTIALS_EXCEPTION = fastapi.HTTPException(
    status_code=fastapi.status.HTTP_401_UNAUTHORIZED,
    detail="Invalid access token.",
    # This header indicates what type of authentication would be required to access
    # the resource.
    # https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
    headers={"WWW-Authenticate": "Bearer"},
)


def optional_token_bearer(request: fastapi.Request):
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


class Permissions:
    def __init__(
        self,
        entry_expiration_min: int = 5,
    ):
        """
        entry_expiration_seconds: number of seconds until the permissions for a
            token needs to be refreshed. We refresh in case a previously invalid
            token becomes a valid token.
        """
        self.general_variables = GeneralVariables.load_from_env()
        self.entry_expiration_min = entry_expiration_min
        self.expirations: List[Tuple[int, str]] = []
        self.cache: Dict[str, dict] = {}
        self.cache_lock = Lock()

    def _clear_expired_entries(self):
        pos = 0
        curr_time = now()
        for expiration, token in self.expirations:
            if expiration > curr_time:
                break
            try:
                del self.cache[token]
            except:
                pass
            pos += 1
        self.expirations = self.expirations[pos:]

    def _get_permissions(self, token) -> Tuple[bool, bool, bool]:
        self._clear_expired_entries()
        curr_time = now()
        if not token in self.cache:
            permissions = self.general_variables.deployment_permissions(token)
            self.expirations.append(
                (
                    curr_time + datetime.timedelta(minutes=self.entry_expiration_min),
                    token,
                )
            )
            self.cache[token] = permissions
            return permissions["read"], permissions["write"], permissions["override"]
        if self.cache[token]["exp"] <= curr_time:
            return False, False, False
        permissions = self.cache[token]
        return permissions["read"], permissions["write"], permissions["override"]

    def verify_read_permission(
        self, token: str = fastapi.Depends(optional_token_bearer)
    ) -> bool:
        with self.cache_lock:
            if not self._get_permissions(token)[0]:
                raise CREDENTIALS_EXCEPTION
        return token

    def verify_write_permission(
        self, token: str = fastapi.Depends(optional_token_bearer)
    ) -> bool:
        with self.cache_lock:
            if not self._get_permissions(token)[1]:
                raise CREDENTIALS_EXCEPTION
        return token

    def get_owner_permission(
        self, token: str = fastapi.Depends(optional_token_bearer)
    ) -> bool:
        with self.cache_lock:
            return self._get_permissions(token)[2]
