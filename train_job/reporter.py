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

    def report_complete(self, model_id: str, metadata: Dict[str, str]):
        """
        Report the completion of a training job.
        Args:
            model_id (str): The ID of the model.
            metadata (Dict[str, str]): Metadata associated with the training job.
        """
        json_data = {
            "model_id": model_id,
            "metadata": metadata,
        }
        content = self._request("post", "api/train/complete", json=json_data)
        print(content)

    def report_status(self, model_id: str, status: str, message: str = ""):
        """
        Report the status of a training job.
        Args:
            model_id (str): The ID of the model.
            status (str): The status of the training job.
            message (str, optional): Additional message. Defaults to "".
        """
        content = self._request(
            "post",
            "api/train/update-status",
            params={"model_id": model_id, "status": status, "message": message},
        )
        print(content)

    def create_shard(
        self,
        shard_num: int,
        model_id: str,
        data_id: str,
        base_model_id: Optional[str] = None,
        extra_options: Optional[Dict] = None,
    ):
        """
        Create a new shard for training.
        Args:
            shard_num (int): The shard number.
            model_id (str): The ID of the model.
            data_id (str): The ID of the data.
            base_model_id (Optional[str], optional): The base model ID. Defaults to None.
            extra_options (Optional[Dict], optional): Extra options for the shard creation. Defaults to None.
        """
        if extra_options is None:
            extra_options = {}
        content = self._request(
            "post",
            "api/train/create-shard",
            params={
                "shard_num": shard_num,
                "model_id": model_id,
                "data_id": data_id,
                "base_model_id": base_model_id,
            },
            data={"extra_options_form": json.dumps(extra_options)},
        )
        print(content)

    def get_model_shard_train_status(self, model_id: str) -> Optional[Dict]:
        """
        Get the training status of a model shard.
        Args:
            model_id (str): The ID of the model.
        Returns:
            Optional[Dict]: The status of the model shard training.
        """
        content = self._request(
            "get", "api/train/model-shard-train-status", params={"model_id": model_id}
        )
        print(content)
        return content

    def report_shard_train_status(
        self, model_id: str, shard_num: int, status: str, message: str = ""
    ):
        """
        Report the training status of a specific shard.
        Args:
            model_id (str): The ID of the model.
            shard_num (int): The shard number.
            status (str): The status of the shard training.
            message (str, optional): Additional message. Defaults to "".
        """
        content = self._request(
            "post",
            "api/train/update-shard-train-status",
            params={
                "shard_num": shard_num,
                "model_id": model_id,
                "status": status,
                "message": message,
            },
        )
        print(content)
