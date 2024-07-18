import datetime
import enum
import os
import re
import shutil
import traceback
from functools import wraps
from pathlib import Path

import requests
from fastapi import UploadFile, status
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from thirdai import neural_db as ndb
from . import logger

def log_function_name(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        logger.info(f"Invoked: {func.__name__}")
        return func(*args, **kwargs)

    return wrapper

def response(status_code: int, message: str, data={}, success: bool = None):
    if success is not None:
        status = "success" if success else "failed"
    else:
        status = "success" if status_code < 400 else "failed"
    return JSONResponse(
        status_code=status_code,
        content={"status": status, "message": message, "data": jsonable_encoder(data)},
    )


def now():
    return datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0)


def delete_job(deployment_id, task_runner_token):
    job_id = f"deployment-{deployment_id}"
    job_url = f"http://172.17.0.1:4646/v1/jobs/{job_id}"
    headers = {"X-Nomad-Token": task_runner_token}
    response = requests.delete(job_url, headers=headers)
    return response, job_id


def propagate_error(func):
    @wraps(func)
    def method(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message=str(traceback.format_exc()),
                success=False,
            )

    return method


def validate_name(name):
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


def validate_files(documents, files: list[UploadFile], data_dir):

    filename_to_file = {file.filename: file for file in files}

    # Ensure that all documents that need to be supplied a file are supplied a file,
    # e.g. local documents that are PDFs, CSVs, DOCXs, need to have a file uploaded
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
                    f"There was an error reading the file from nfs : {error}"
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

            except Exception:
                raise Exception(
                    f"There was an error uploading the file from local: {error}"
                )


class Status(str, enum.Enum):
    not_started = "not_started"
    starting = "starting"
    in_progress = "in_progress"
    stopping = "stopping"
    complete = "complete"
    failed = "failed"


def create_ndb_docs(documents, data_dir):

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
                        aws_access_key_id=os.getenv("AWS_ACCESS_KEY"),
                        aws_secret_access_key=os.getenv("AWS_ACCESS_SECRET"),
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
                    f"There was an error downloading the file from s3 : {error}"
                )

        if doc["document_type"] in FILE_DOCUMENT_TYPES:
            file_path = os.path.join(data_dir, os.path.basename(doc["path"]))
            ndb_doc_params = {
                **(
                    {
                        key: value
                        for key, value in doc.items()
                        if key != "document_type"
                        and key != "location"
                        and key != "path"
                    }
                )
            }
            ndb_doc = getattr(ndb, doc["document_type"])(file_path, **ndb_doc_params)

        else:
            ndb_doc_params = {
                **(
                    {
                        key: value
                        for key, value in doc.items()
                        if key != "document_type" and key != "location"
                    }
                )
            }
            ndb_doc = getattr(ndb, doc["document_type"])(**ndb_doc_params)

        if doc["location"] == "s3":
            ndb_doc.path = Path(f"/{bucket}.s3.amazonaws.com/{object}")
            os.remove(os.path.join(data_dir, os.path.basename(doc["path"])))
        ndb_docs.append(ndb_doc)

    return ndb_docs
