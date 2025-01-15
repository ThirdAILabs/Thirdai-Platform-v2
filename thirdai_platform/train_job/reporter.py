from abc import ABC, abstractmethod
from typing import Dict, Optional
from urllib.parse import urljoin

import requests
from platform_common.logging import JobLogger


class Reporter(ABC):
    @abstractmethod
    def report_complete(self, model_id: str, metadata: Dict[str, str]):
        raise NotImplementedError

    @abstractmethod
    def report_status(self, model_id: str, status: str, message: Optional[str] = None):
        raise NotImplementedError

    @abstractmethod
    def report_warning(self, model_id: str, message: str):
        raise NotImplementedError


class HttpReporter(Reporter):
    def __init__(self, api_url: str, auth_token: str, logger: JobLogger):
        """
        Initialize the Reporter with the given API URL.
        """
        self._api = api_url
        self._auth_token = auth_token
        self.logger = logger

    def _request(self, method: str, suffix: str, *args, **kwargs) -> Optional[Dict]:
        """
        Make an HTTP request with the given method and URL suffix.
        Args:
            method (str): HTTP method (e.g., 'post', 'get').
            suffix (str): URL suffix to append to the base API URL.
            *args: Additional positional arguments for the request.
            **kwargs: Additional keyword arguments for the request.
        Returns:
            Optional[Dict]: Parsed JSON response content if successful, None otherwise.
        """
        # Add custom user-agent to avoid ngrok abuse page
        if "headers" not in kwargs:
            kwargs["headers"] = {}

        kwargs["headers"].update(
            {"Authorization": f"Bearer {self._auth_token}", "User-Agent": "Train job"}
        )

        url = urljoin(self._api, suffix)

        self.logger.info(f"Making {method.upper()} request to {url}")
        self.logger.debug(f"Request headers: {kwargs.get('headers')}")
        if "json" in kwargs:
            self.logger.debug(f"Request JSON payload: {kwargs['json']}")
        if "params" in kwargs:
            self.logger.debug(f"Request query parameters: {kwargs['params']}")
        try:
            response = requests.request(method, url, *args, **kwargs)
            response.raise_for_status()
            self.logger.debug(f"Response content: {response.json()}")
            return response.json()
        except requests.exceptions.RequestException as exception:
            self.logger.error(f"Request to {url} failed with error: {exception}")
            raise exception

    def report_complete(self, model_id: str, metadata: Dict[str, str]):
        """
        Report the completion of a training job.
        Args:
            model_id (str): The ID of the model.
            metadata (Dict[str, str]): Metadata associated with the training job.
        """
        json_data = {
            "status": "complete",
            "metadata": metadata,
        }
        self._request("post", "api/v2/train/update-status", json=json_data)

    def report_status(self, model_id: str, status: str, message: str = ""):
        """
        Report the status of a training job.
        Args:
            model_id (str): The ID of the model.
            status (str): The status of the training job.
            message (str, optional): Additional message. Defaults to "".
        """
        if status == "failed":
            self.report_error(message)

        self._request("post", "api/v2/train/update-status", json={"status": status})

    def report_error(self, message: str):
        self._request(
            "post",
            "api/v2/train/log",
            json={"level": "error", "message": message},
        )

    def report_warning(self, model_id: str, message: str):
        self._request(
            "post",
            "api/v2/train/log",
            json={"level": "warning", "message": message},
        )
