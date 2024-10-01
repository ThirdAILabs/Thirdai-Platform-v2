import json
import sys
import time
from functools import wraps
from urllib.parse import urljoin

import requests
from IPython.display import clear_output


def print_progress_dots(duration: int):
    for _ in range(duration):
        sys.stdout.write(".")
        sys.stdout.flush()
        time.sleep(1)
    clear_output(wait=True)


def create_model_identifier(model_name: str, author_username: str):
    return author_username + "/" + model_name


def construct_deployment_url(host, model_id):
    return urljoin(host, model_id) + "/"


def check_deployment_decorator(func):
    """
    A decorator function to check if deployment is complete before executing the decorated method.

    Args:
        func (callable): The function to be decorated.

    Returns:
        callable: The decorated function.
    """

    @wraps(func)
    def wrapper(self, *args, **kwargs):
        try:
            return func(self, *args, **kwargs)
        except requests.RequestException as e:
            print(f"Error during HTTP request: {str(e)}")
            print(
                "Deployment might not be complete yet. Call `list_deployments()` to check status of your deployment."
            )
            return None

    return wrapper


def check_response(response):
    if not (200 <= response.status_code < 300):
        print(response.content)
        raise requests.exceptions.HTTPError(
            "Failed with status code:", response.status_code, response=response
        )

    content = json.loads(response.content)
    print(content)

    status = content["status"]

    if status != "success":
        error = content["message"]
        raise requests.exceptions.HTTPError(f"error: {error}")


def http_get_with_error(*args, **kwargs):
    """Makes an HTTP GET request and raises an error if status code is not
    2XX.
    """
    response = requests.get(*args, **kwargs)
    print("Response GET:", response.json())
    check_response(response)
    return response


def http_post_with_error(*args, **kwargs):
    """Makes an HTTP POST request and raises an error if status code is not
    2XX.
    """
    response = requests.post(*args, **kwargs)
    print("Response POST:", response.json())
    check_response(response)
    return response


def http_delete_with_error(*args, **kwargs):
    """Makes an HTTP POST request and raises an error if status code is not
    2XX.
    """
    response = requests.delete(*args, **kwargs)
    check_response(response)
    return response


def auth_header(access_token):
    return {
        "Authorization": f"Bearer {access_token}",
    }
