import json
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

        kwargs["headers"].update({"User-Agent": "NDB Deployment job"})

        url = urljoin(self._api, suffix)
        try:
            if method == "post":
                response = requests.post(url, *args, **kwargs)
            elif method == "get":
                response = requests.get(url, *args, **kwargs)
            if 200 <= response.status_code < 300:
                content = json.loads(response.content)
                return content
            else:
                print(response.content)
                raise requests.exceptions.HTTPError(
                    "Failed with status code:", response.status_code
                )

        except Exception as exception:
            # This could be other forms of error like connection error, stuff not found.
            # In this case, we log the response from the server, but ignore the error.
            print(exception)
            raise exception

    def save_model(
        self,
        access_token,
        deployment_id,
        model_id,
        base_model_id,
        model_name,
        metadata,
    ):
        content = self._request(
            "post",
            f"api/model/save-deployed",
            json={
                "deployment_id": deployment_id,
                "model_id": model_id,
                "base_model_id": base_model_id,
                "model_name": model_name,
                "metadata": metadata,
            },
            headers=self.auth_header(access_token=access_token),
        )
        print(content)

    def auth_header(self, access_token):
        return {
            "Authorization": f"Bearer {access_token}",
        }

    def check_model_present(
        self,
        access_token,
        model_name,
    ):
        content = self._request(
            "get",
            f"api/model/name-check",
            params={
                "name": model_name,
            },
            headers=self.auth_header(access_token=access_token),
        )
        print(content)

        return content["data"]["model_present"]

    def deploy_complete(self, deployment_id):
        content = self._request(
            "post",
            f"api/deploy/complete",
            params={
                "deployment_id": deployment_id,
            },
        )
        print(content)

    def deploy_log(
        self, action, deployment_id, train_samples, access_token, used=False
    ):
        content = self._request(
            "post",
            f"api/deploy/log",
            json={
                "deployment_id": deployment_id,
                "action": action,
                "train_samples": train_samples,
                "used": used,
            },
            headers=self.auth_header(access_token=access_token),
        )

        print(content)

    def action_log(self, action, train_samples, used=False, **kwargs):
        params = (
            {
                "action": action,
                "train_samples": train_samples,
                "used": used,
            },
        )
        params.update(**kwargs)
        content = self._request(
            "post",
            f"api/logger/log",
            json=params,
        )

        print(content)
