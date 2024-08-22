import json
from typing import Dict, Optional
from urllib.parse import urljoin

import requests


class Reporter:
    def __init__(self, api_url: str):
        """
        Initialize the Reporter with the given API URL.
        """
        self._api = api_url

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

        kwargs["headers"].update({"User-Agent": "Train job"})

        url = urljoin(self._api, suffix)
        try:
            response = requests.request(method, url, *args, **kwargs)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as exception:
            print(exception)
            raise exception

    def update_status(
        self,
        data_id: str,
        status: str,
    ):
        content = self._request(
            "post",
            "api/data/update-catalog",
            params={
                "data_id": data_id,
                "status": status,
            },
        )
        print(content)

    def report_generate_status(
        self, data_id: str, name: int, task: str, target_labels: list[str] = []
    ):
        """
        Report the training status of a specific shard.
        Args:
            data (str): The ID catalog.
            name (int): name of dataset.
            task (str): udt task.
            target_labels (List[str], optional): list of tags/labels.
        """
        content = self._request(
            "post",
            "api/data/create-catalog",
            params={
                "data_id": data_id,
                "name": name,
                "task": task,
                "target_labels": target_labels,
            },
        )
        print(content)
