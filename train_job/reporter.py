import json
from typing import Dict
from urllib.parse import urljoin

import requests


class Reporter:
    def __init__(self, api_url):
        self._api = api_url

    def _request(self, method, suffix, *args, **kwargs):
        # The following exists to have custom user-agent so ngrok doesn't
        # provide an abuse page.
        if "headers" not in kwargs:
            kwargs["headers"] = {}

        kwargs["headers"].update({"User-Agent": "Train job"})

        url = urljoin(self._api, suffix)
        try:
            if method == "post":
                response = requests.post(url, *args, **kwargs)
            elif method == "get":
                response = requests.get(url, *args, **kwargs)
            if response.status_code == 200:
                content = json.loads(response.content)
                return content

            print(response)

        except Exception as exception:
            # This could be other forms of error like connection error, stuff not found.
            # In this case, we log the response from the server, but ignore the error.
            print(exception)
            raise exception

    def report_complete(
        self,
        model_id: str,
        metadata: Dict[str, str],
    ):
        json_data = {
            "model_id": model_id,
            "metadata": metadata,
        }
        content = self._request(
            "post",
            f"api/train/complete",
            json=json_data,
        )
        print(content)

    def report_status(self, model_id, status, message=""):
        content = self._request(
            "post",
            f"api/train/update-status",
            params={"model_id": model_id, "status": status, "message": message},
        )
        print(content)

    def create_shard(
        self,
        shard_num,
        model_id,
        data_id,
        base_model_id=None,
        extra_options={},
    ):
        content = self._request(
            "post",
            f"api/train/create-shard",
            params={
                "shard_num": shard_num,
                "model_id": model_id,
                "data_id": data_id,
                "base_model_id": base_model_id,
            },
            data={"extra_options_form": json.dumps(extra_options)},
        )
        print(content)

    def get_model_shard_train_status(self, model_id):
        content = self._request(
            "get",
            f"api/train/model-shard-train-status",
            params={"model_id": model_id},
        )
        print(content)
        return content

    def report_shard_train_status(self, model_id, shard_num, status, message=""):
        content = self._request(
            "post",
            f"api/train/update-shard-train-status",
            params={
                "shard_num": shard_num,
                "model_id": model_id,
                "status": status,
                "message": message,
            },
        )
        print(content)
