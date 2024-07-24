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


def validate_files(
    documents: List[Dict[str, Any]], files: List[UploadFile], data_dir: str
) -> None:
    """
    Validates that all required files are provided for the documents and saves them to the data directory.

    Args:
        documents (List[Dict[str, Any]]): List of document dictionaries.
        files (List[UploadFile]): List of uploaded files.
        data_dir (str): Directory to save the files.

    Raises:
        Exception: If there is a mismatch between documents and uploaded files or an error during file handling.
    """
    filename_to_file = {file.filename: file for file in files}

    filenames = set(filename_to_file.keys())
    file_doc_names = set(
        [
            os.path.basename(doc["path"])
            for doc in documents
            if doc["location"] == "local"
            and doc["document_type"] in FILE_DOCUMENT_TYPES
        ]
    )
    if filenames != file_doc_names:
        raise Exception("Mismatch between documents and uploaded files")

    for doc in documents:
        if doc["location"] == "nfs":
            try:
                shutil.copy(doc["path"], data_dir)
            except Exception as error:
                raise Exception(
                    f"There was an error reading the file from nfs: {error}"
                )

        if doc["location"] == "local":
            try:
                if doc["document_type"] in FILE_DOCUMENT_TYPES:
                    file = filename_to_file[os.path.basename(doc["path"])]
                    contents = file.file.read()
                    with open(
                        os.path.join(data_dir, os.path.basename(doc["path"])), "wb"
                    ) as f:
                        f.write(contents)

            except Exception as error:
                raise Exception(
                    f"There was an error uploading the file from local: {error}"
                )


class Status(str, enum.Enum):
    not_started = "not_started"
    starting = "starting"
    in_progress = "in_progress"
    stopped = "stopped"
    complete = "complete"
    failed = "failed"


def create_ndb_docs(
    documents: List[Dict[str, Any]], data_dir: str
) -> List[ndb.Document]:
    """
    Creates NDB documents from the provided document dictionaries.

    Args:
        documents (List[Dict[str, Any]]): List of document dictionaries.
        data_dir (str): Directory to save the files.

    Returns:
        List[ndb.Document]: List of NDB documents.
    """
    ndb_docs = []

    for doc in documents:
        if doc["location"] == "s3":
            import boto3
            from botocore import UNSIGNED
            from botocore.client import Config

            def create_s3_client():
                aws_access_key = os.getenv("AWS_ACCESS_KEY")
                aws_secret_access_key = os.getenv("AWS_ACCESS_SECRET")
                if not aws_access_key or not aws_secret_access_key:
                    config = Config(
                        signature_version=UNSIGNED,
                        retries={"max_attempts": 10, "mode": "standard"},
                        connect_timeout=5,
                        read_timeout=60,
                    )
                    s3_client = boto3.client(
                        "s3",
                        config=config,
                    )
                else:
                    config = Config(
                        retries={"max_attempts": 10, "mode": "standard"},
                        connect_timeout=5,
                        read_timeout=60,
                    )
                    s3_client = boto3.client(
                        "s3",
                        aws_access_key_id=aws_access_key,
                        aws_secret_access_key=aws_secret_access_key,
                        config=config,
                    )
                return s3_client

            s3 = create_s3_client()
            bucket, object = doc["path"].replace("s3://", "").split("/", 1)
            try:
                s3.download_file(
                    bucket,
                    object,
                    os.path.join(data_dir, os.path.basename(doc["path"])),
                )
            except Exception as error:
                raise Exception(
                    f"There was an error downloading the file from s3: {error}"
                )

        if doc["document_type"] in FILE_DOCUMENT_TYPES:
            file_path = os.path.join(data_dir, os.path.basename(doc["path"]))
            ndb_doc_params = {
                key: value
                for key, value in doc.items()
                if key not in {"document_type", "location", "path"}
            }
            ndb_doc = getattr(ndb, doc["document_type"])(file_path, **ndb_doc_params)

        else:
            ndb_doc_params = {
                key: value
                for key, value in doc.items()
                if key not in {"document_type", "location"}
            }
            ndb_doc = getattr(ndb, doc["document_type"])(**ndb_doc_params)

        if doc["location"] == "s3":
            ndb_doc.path = Path(f"/{bucket}.s3.amazonaws.com/{object}")
            os.remove(os.path.join(data_dir, os.path.basename(doc["path"])))
        ndb_docs.append(ndb_doc)

    return ndb_docs
