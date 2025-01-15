from typing import Optional
from urllib.parse import urljoin

import requests
from platform_common.logging import JobLogger


class Reporter:
    def __init__(self, api_url: str, auth_token: str, logger: JobLogger):
        """
        Initializes the Reporter instance with the API URL.

        Args:
            api_url (str): The base URL for the API.
        """
        self._api = api_url
        self._auth_token = auth_token
        self.logger = logger

    def _request(self, method: str, suffix: str, *args, **kwargs) -> dict:
        """
        Makes an HTTP request to the specified API endpoint.

        Args:
            method (str): The HTTP method to use ('post' or 'get').
            suffix (str): The API endpoint suffix.
            *args: Additional positional arguments for the request.
            **kwargs: Additional keyword arguments for the request.

        Returns:
            dict: The JSON response content.

        Raises:
            requests.exceptions.HTTPError: If the request fails with an HTTP error.
            Exception: For other types of exceptions.
        """
        if "headers" not in kwargs:
            kwargs["headers"] = {}

        # The following exists to have custom user-agent so ngrok doesn't
        # provide an abuse page.
        kwargs["headers"]["User-Agent"] = "NDB Deployment job"

        if not "Authorization" in kwargs["headers"]:
            kwargs["headers"]["Authorization"] = f"Bearer {self._auth_token}"

        url = urljoin(self._api, suffix)

        self.logger.info(
            f"Making {method.upper()} request to {url} with args: {args}, kwargs: {kwargs}"
        )
        try:
            response = requests.request(method, url, *args, **kwargs)
            response.raise_for_status()
            content = response.json()
            self.logger.info(f"Response from {url}: {content}")
            return content
        except requests.exceptions.HTTPError as http_err:
            self.logger.error(
                f"HTTPError for {method.upper()} request to {url}: {http_err}, Response: {response.text}"
            )
            raise
        except Exception as e:
            self.logger.error(f"Error during {method.upper()} request to {url}: {e}")
            raise

    def save_model(self, access_token: str, base_model_id: str, model_name: str) -> str:
        """
        Saves the deployed model information.

        Args:
            access_token (str): The access token for authentication.
            model_id (str): The ID of the model.
            base_model_id (str): The ID of the base model.
            model_name (str): The name of the model.
            metadata (dict): Metadata associated with the model.
        """
        return self._request(
            "post",
            f"api/v2/deploy/{base_model_id}/save",
            json={"model_name": model_name},
            headers=self.auth_header(access_token=access_token),
        )

    def save_complete(self, token: str):
        self._request(
            "post",
            f"api/v2/train/update-status",
            json={"status": "complete"},
            headers=self.auth_header(token),
        )

    def auth_header(self, access_token: str) -> dict:
        """
        Generates the authentication header.

        Args:
            access_token (str): The access token for authentication.

        Returns:
            dict: The authentication header.
        """
        return {
            "Authorization": f"Bearer {access_token}",
        }

    def update_deploy_status(
        self, model_id: str, status: str, message: Optional[str] = None
    ) -> None:
        """
        Updates the deployment status.

        Args:
            model_id (str): The ID of the model.
            status (str): The new status of the deployment.
        """
        self._request("post", "api/v2/deploy/update-status", json={"status": status})

    def get_deploy_status(self, model_id: str) -> str:
        """
        Gets the deployment status.

        Args:
            model_id (str): The ID of the model.
        """
        content = self._request("get", "api/v2/deploy/status-internal")
        return content["status"]

    def log(
        self,
        action: str,
        model_id: str,
        train_samples: list,
        access_token: str,
        used: bool = False,
    ) -> None:
        """
        Logs an action for the deployment.

        Args:
            action (str): The action to log.
            model_id (str): The ID of the model.
            train_samples (list): The training samples associated with the action.
            access_token (str): The access token for authentication.
            used (bool): Whether the action was used. Defaults to False.
        """
        self._request(
            "post",
            "api/deploy/log",
            json={
                "model_id": model_id,
                "action": action,
                "train_samples": train_samples,
                "used": used,
            },
            headers=self.auth_header(access_token=access_token),
        )
