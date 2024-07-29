import datetime
import enum
import os
import re
import shutil
import traceback
from functools import wraps
from pathlib import Path
from typing import Any, Dict, List, Tuple

import requests
from fastapi import UploadFile, status
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from thirdai import neural_db as ndb


def response(
    status_code: int, message: str, data: Dict = {}, success: bool = None
) -> JSONResponse:
    """
    Creates a JSON response with a given status code, message, and data.

    Args:
        status_code (int): HTTP status code.
        message (str): Message to include in the response.
        data (Dict, optional): Data to include in the response. Defaults to {}.
        success (bool, optional): Indicates success or failure. Defaults to None.

    Returns:
        JSONResponse: The JSON response.
    """
    if success is not None:
        status = "success" if success else "failed"
    else:
        status = "success" if status_code < 400 else "failed"
    return JSONResponse(
        status_code=status_code,
        content={"status": status, "message": message, "data": jsonable_encoder(data)},
    )


def now() -> datetime.datetime:
    """
    Returns the current UTC time without microseconds.

    Returns:
        datetime.datetime: Current UTC time.
    """
    return datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0)


def delete_job(
    deployment_id: str, task_runner_token: str
) -> Tuple[requests.Response, str]:
    """
    Deletes a job from Nomad.

    Args:
        deployment_id (str): The deployment ID.
        task_runner_token (str): The task runner token.

    Returns:
        Tuple[requests.Response, str]: The response from the delete request and the job ID.
    """
    job_id = f"deployment-{deployment_id}"
    job_url = f"http://172.17.0.1:4646/v1/jobs/{job_id}"
    headers = {"X-Nomad-Token": task_runner_token}
    response = requests.delete(job_url, headers=headers)
    return response, job_id


def propagate_error(func):
    """
    Decorator to propagate errors and return a JSON response with the error message.

    Args:
        func: The function to wrap.

    Returns:
        The wrapped function.
    """

    @wraps(func)
    def method(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception:
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message=str(traceback.format_exc()),
                success=False,
            )

    return method


def validate_name(name: str) -> None:
    """
    Validates a name to ensure it matches a specific regex pattern.

    Args:
        name (str): The name to validate.

    Raises:
        ValueError: If the name does not match the regex pattern.
    """
    regex_pattern = "^[\w-]+$"
    if not re.match(regex_pattern, name):
        raise ValueError("name is not valid")


FILE_DOCUMENT_TYPES = [
    "CSV",
    "PDF",
    "DOCX",
    "SentenceLevelPDF",
    "SentenceLevelDOCX",
    "Unstructured",
]


class Status(str, enum.Enum):
    not_started = "not_started"
    starting = "starting"
    in_progress = "in_progress"
    stopped = "stopped"
    complete = "complete"
    failed = "failed"
