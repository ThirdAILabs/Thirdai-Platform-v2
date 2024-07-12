import json
import logging
import os
import re
from enum import Enum
from pathlib import Path
from typing import Dict, List, Optional
from urllib.parse import urljoin

import bcrypt
import requests
from database import schema
from fastapi import UploadFile
from fastapi.responses import JSONResponse
from jinja2 import Template
from pydantic import BaseModel, Field, validator
from sqlalchemy.orm import Session

logger = logging.getLogger("ThirdAI_Platform")


def setup_logger(
    level=logging.DEBUG, format="%(asctime)s | [%(name)s] [%(levelname)s] %(message)s"
):
    console_handler = logging.StreamHandler()
    console_handler.setLevel(level)

    formatter = logging.Formatter(format)
    console_handler.setFormatter(formatter)

    # add ch to logger
    logger.addHandler(console_handler)

    logger.setLevel(level)
    logger.info("Initialized console logging.")


def response(status_code: int, message: str, data={}, success: bool = None):
    if success is not None:
        status = "success" if success else "failed"
    else:
        status = "success" if status_code < 400 else "failed"
    return JSONResponse(
        status_code=status_code,
        content={"status": status, "message": message, "data": data},
    )


def hash_password(password: str):
    byte_password = password.encode("utf-8")
    salt = bcrypt.gensalt()
    return bcrypt.hashpw(byte_password, salt).decode()


def get_high_level_model_info(result: schema.Model):
    info = {
        "model_name": result.name,
        "publish_date": str(result.published_date),
        "user_email": result.user.email,
        "username": result.user.username,
        "access_level": result.access_level,
        "domain": result.domain,
    }

    # Include metadata if it exists
    if result.meta_data:
        metadata = result.meta_data
        if metadata.public:
            info.update(metadata.public)
        if metadata.protected:
            info.update(metadata.protected)
        if metadata.private:
            info.update(metadata.private)

    return info


def validate_name(name):
    regex_pattern = "^[\w-]+$"
    if not re.match(regex_pattern, name):
        raise ValueError("name is not valid")


class NDBExtraOptions(BaseModel):
    # ----shard specific training params----
    num_models_per_shard: Optional[int] = Field(1, gt=0)
    num_shards: Optional[int] = Field(1, gt=0)
    allocation_cores: Optional[int] = None
    allocation_memory: Optional[int] = None

    # ----shard agnostic training params----
    model_cores: Optional[int] = None
    model_memory: Optional[int] = None
    priority: Optional[int] = None
    csv_id_column: Optional[str] = None
    csv_strong_columns: Optional[list[str]] = None
    csv_weak_columns: Optional[list[str]] = None
    csv_reference_columns: Optional[list[str]] = None
    fhr: Optional[int] = None
    embedding_dim: Optional[int] = None
    output_dim: Optional[int] = None
    max_in_memory_batches: Optional[int] = None
    extreme_num_hashes: Optional[int] = None
    num_classes: Optional[int] = None

    csv_query_column: Optional[str] = None
    csv_id_delimiter: Optional[str] = None

    learning_rate: Optional[float] = None
    batch_size: Optional[int] = None
    unsupervised_epochs: Optional[int] = None
    supervised_epochs: Optional[int] = None

    tokenizer: Optional[str] = None
    hidden_bias: Optional[bool] = None
    retriever: Optional[str] = None  # This flag is for which retriever to use.
    unsupervised_train: Optional[bool] = None
    disable_finetunable_retriever: Optional[bool] = (
        None  # This flag is to disable inverted index in supervised training.
    )

    class Config:
        extra = "forbid"


class FileType(str, Enum):
    unsupervised = "unsupervised"
    supervised = "supervised"
    test = "test"


class FileLocation(str, Enum):
    local = "local"
    nfs = "nfs"
    s3 = "s3"


class FileDetails(BaseModel):
    mode: FileType
    location: FileLocation
    source_id: Optional[str] = None
    metadata: Optional[Dict[str, str]] = None

    is_folder: Optional[bool] = False

    @validator("location")
    def check_location(cls, v):
        if v not in FileLocation.__members__.values():
            raise ValueError(
                f"Invalid location value. Supported locations are {list(FileLocation)}"
            )
        return v

    @validator("source_id", always=True)
    def check_source_id(cls, v, values):
        if values.get("mode") == FileType.supervised and not v:
            raise ValueError("source_id is required for supervised files.")
        if values.get("mode") != FileType.supervised and v:
            raise ValueError("source_id is only allowed for supervised files.")
        if values.get("source_id") and v is not None:
            raise ValueError("source_id should not be provided when is_folder is True")
        return v

    @validator("metadata", always=True)
    def check_metadata(cls, v, values):
        if values.get("is_folder") and v is not None:
            raise ValueError("metadata should not be provided when is_folder is True")
        if values.get("mode") == FileType.supervised and v is not None:
            raise ValueError(
                "metadata is not required for supervised files, it is only for unsupervised files."
            )
        return v

    @validator("mode", always=True)
    def check_unsupervised_metadata(cls, v, values):
        if v == FileType.unsupervised and values.get("metadata") is not None:
            if values.get("is_folder"):
                raise ValueError(
                    "metadata should not be provided when is_folder is True"
                )
        if values.get("is_folder") and v != FileType.unsupervised:
            raise ValueError("is_folder can only be True for unsupervised files.")
        return v

    @validator("is_folder", always=True)
    def check_is_folder(cls, v, values):
        if v and values.get("location") == FileLocation.local:
            raise ValueError("is_folder can only be True for nfs and s3 locations.")
        if v and values.get("mode") != FileType.unsupervised:
            raise ValueError("is_folder can only be True for unsupervised files.")
        return v

    def validate_csv_extension(self, filename: str):
        if self.mode in {FileType.supervised, FileType.test}:
            _, ext = os.path.splitext(filename)
            if ext != ".csv":
                raise ValueError(
                    f"{filename} file has to be a csv file but given {ext} file."
                )
        return True


class FileDetailsList(BaseModel):
    file_details: List[FileDetails]

    @validator("file_details")
    def check_file_counts(cls, v):
        test_count = sum(1 for file in v if file.mode == FileType.test)
        unsupervised_count = sum(1 for file in v if file.mode == FileType.unsupervised)
        unsupervised_metadata_count = sum(
            1
            for file in v
            if file.mode == FileType.unsupervised and file.metadata is not None
        )

        if test_count > 1:
            raise ValueError("Currently supports a single test file")

        if 0 < unsupervised_metadata_count < unsupervised_count:
            raise ValueError(
                "Either all unsupervised files must have metadata or none should have metadata."
            )

        return v


def get_model(session: Session, username: str, model_name: str):
    return (
        session.query(schema.Model)
        .join(schema.User)
        .filter(schema.User.username == username, schema.Model.name == model_name)
        .first()
    )


def create_s3_client():
    import boto3
    from botocore import UNSIGNED
    from botocore.client import Config

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


def list_files_in_s3(bucket_name, prefix):
    s3 = create_s3_client()

    paginator = s3.get_paginator("list_objects_v2")
    pages = paginator.paginate(Bucket=bucket_name, Prefix=prefix)

    file_keys = []

    for page in pages:
        if "Contents" in page:
            for obj in page["Contents"]:
                file_keys.append(obj["Key"])

    return file_keys


def list_s3_files(filename):
    s3_urls = []
    bucket_name, prefix = filename.replace("s3://", "").split("/", 1)
    file_keys = list_files_in_s3(bucket_name, prefix)
    s3_urls.extend([f"s3://{bucket_name}/{key}" for key in file_keys])

    return s3_urls


def get_files(
    files: List[UploadFile],
    data_id,
    files_info: List[FileDetails],
):
    filenames = []
    supervised_filenames = []
    source_ids = []

    for i, file in enumerate(files):
        file_info = files_info[i]
        destination_dir = os.path.join(
            os.getenv("LOCAL_TEST_DIR", "/model_bazaar"),
            "data",
            str(data_id),
            file_info.mode,
        )
        os.makedirs(destination_dir, exist_ok=True)

        if file_info.mode == FileType.supervised:
            supervised_filenames.append(file.filename)
            source_ids.append(file_info.source_id)

        if file_info.location == FileLocation.nfs:
            nfs_file_path = os.path.join(destination_dir, "nfs_files.txt")
            try:
                if file_info.is_folder:
                    for root, _, files_in_dir in os.walk(file.filename):
                        for filename in files_in_dir:
                            src_file_path = os.path.join(root, filename)
                            with open(nfs_file_path, "a") as nfs_file:
                                nfs_file.write(src_file_path + "\n")
                            filenames.append(src_file_path)
                            file_info.validate_csv_extension(src_file_path)
                else:
                    with open(nfs_file_path, "a") as nfs_file:
                        nfs_file.write(file.filename + "\n")
                    filenames.append(file.filename)
                    file_info.validate_csv_extension(file.filename)
                    if files_info[i].metadata:
                        metadata_file_path = (
                            f"{os.path.splitext(file.filename)[0]}_metadata.json"
                        )
                        with open(metadata_file_path, "w") as json_file:
                            json.dump(files_info[i].metadata, json_file)
            except Exception as error:
                return f"There was an error reading the file from nfs : {error}"
        elif file_info.location == FileLocation.s3:
            s3_file_path = os.path.join(destination_dir, "s3_files.txt")
            try:
                s3_files = list_s3_files(file.filename)
                for s3_file in s3_files:
                    with open(s3_file_path, "a") as s3_file_local:
                        s3_file_local.write(s3_file + "\n")
                    filenames.append(s3_file)
                    file_info.validate_csv_extension(s3_file)
            except Exception as error:
                return f"There was an error writing the S3 URL to the file: {error}"
        else:
            destination_path = os.path.join(
                destination_dir, str(os.path.basename(file.filename))
            )
            os.makedirs(os.path.dirname(destination_path), exist_ok=True)

            if file_info.metadata:
                metadata_file_path = (
                    f"{os.path.splitext(destination_path)[0]}_metadata.json"
                )
                with open(metadata_file_path, "w") as json_file:
                    json.dump(file_info.metadata, json_file)

            file_info.validate_csv_extension(file.filename)

            try:
                contents = file.file.read()
                with open(destination_path, "wb") as f:
                    f.write(contents)
            except Exception as error:
                return f"There was an error uploading the file from local: {error}"
            finally:
                file.file.close()

            filenames.append(file.filename)

    if len(supervised_filenames) > 0:
        save_file_relations(
            supervised_file_names=supervised_filenames,
            data_id=data_id,
            source_ids=source_ids,
        )

    return filenames


def save_file_relations(
    supervised_file_names,
    data_id,
    source_ids,
):
    if len(supervised_file_names) != len(source_ids):
        return {
            "status_code": 400,
            "message": "Source ids have not been given for all supervised files.",
        }

    relations_dict_list = [
        {
            "supervised_file": os.path.basename(file_name),
            "source_id": source_ids[i],
        }
        for i, file_name in enumerate(supervised_file_names)
    ]

    destination_path = os.path.join(
        os.getenv("LOCAL_TEST_DIR", "/model_bazaar"),
        "data",
        str(data_id),
        "relations.json",
    )

    os.makedirs(os.path.dirname(destination_path), exist_ok=True)

    with open(destination_path, "w") as file:
        json.dump(relations_dict_list, file, indent=4)


def parse_model_identifier(model_identifier):
    regex_pattern = "^[\w-]+\/[\w-]+$"
    if re.match(regex_pattern, model_identifier):
        username, model_name = model_identifier.split("/")
        return username, model_name
    else:
        raise ValueError("model identifier is not valid")


def get_model_from_identifier(model_identifier, session):
    try:
        model_username, model_name = parse_model_identifier(model_identifier)
    except Exception as error:
        raise ValueError(str(error))
    model: schema.Model = get_model(
        session, username=model_username, model_name=model_name
    )
    if not model:
        raise ValueError("There is no model with the given name.")
    return model


TASK_RUNNER_TOKEN = os.getenv("TASK_RUNNER_TOKEN")


def get_hcl_payload(filepath, is_jinja, **kwargs):
    with open(filepath, "r") as file:
        content = file.read()

    if is_jinja:
        template = Template(content)
        hcl_content = template.render(**kwargs)
    else:
        hcl_content = content

    payload = {"JobHCL": hcl_content, "Canonicalize": True}

    return payload


def nomad_job_exists(job_id, nomad_endpoint):
    headers = {"X-Nomad-Token": TASK_RUNNER_TOKEN}
    response = requests.get(
        urljoin(nomad_endpoint, f"v1/job/{job_id}"), headers=headers
    )
    return response.status_code == 200


def submit_nomad_job(filepath, nomad_endpoint, **kwargs):
    """
    This function submits an generated HCL job file from a jinja file to nomad
    """

    json_payload_url = urljoin(nomad_endpoint, "v1/jobs/parse")
    submit_url = urljoin(nomad_endpoint, "v1/jobs")
    headers = {"Content-Type": "application/json", "X-Nomad-Token": TASK_RUNNER_TOKEN}

    filepath_ext = filepath.split(".")[-1]
    is_jinja = filepath_ext == "j2"
    hcl_payload = get_hcl_payload(filepath, is_jinja=is_jinja, **kwargs)

    # Before submitting a job to nomad, we must convert the HCL file to JSON
    json_payload_response = requests.post(
        json_payload_url, headers=headers, json=hcl_payload
    )
    json_payload = json_payload_response.json()

    # Submit the JSON job spec to Nomad
    response = requests.post(submit_url, headers=headers, json={"Job": json_payload})

    return response


def delete_nomad_job(job_id, nomad_endpoint):
    job_url = urljoin(nomad_endpoint, f"v1/job/{job_id}")
    headers = {"X-Nomad-Token": TASK_RUNNER_TOKEN}
    response = requests.delete(job_url, headers=headers)

    if response.status_code == 200:
        print(f"Job {job_id} stopped successfully")
    else:
        print(
            f"Failed to stop job {job_id}. Status code: {response.status_code}, Response: {response.text}"
        )

    return response


def get_platform():
    platform = os.getenv("PLATFORM", "docker")
    options = ["docker", "local"]
    if platform not in options:
        print(
            f"Invalid platform identifier '{platform}'. Options: {options}. Defaulting to docker."
        )
    return platform


def get_python_path():
    python_path = "python3"
    if get_platform() == "local":
        python_path = os.getenv("PYTHON_PATH")
        if not python_path:
            raise ValueError(
                "PYTHON_PATH environment variable is not set for local development."
            )

    return python_path


def get_root_absolute_path():
    return Path(__file__).parent.parent.parent.absolute()


def update_json(current_json, new_dict):
    if current_json is None:
        current_dict = {}
    else:
        current_dict = json.loads(current_json)
    current_dict.update(new_dict)
    return json.dumps(current_dict)
