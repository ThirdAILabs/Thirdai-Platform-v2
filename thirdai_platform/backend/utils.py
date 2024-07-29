import json
import logging
import os
import re
import socket
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
from pydantic import BaseModel, Field, root_validator, validator
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
        "type": result.type,
    }

    # Include metadata if it exists
    if result.meta_data:
        metadata = result.meta_data
        if metadata.train:
            info.update(metadata.train)
        if metadata.general:
            info.update(metadata.general)

    return info


def validate_name(name):
    regex_pattern = "^[\w-]+$"
    if not re.match(regex_pattern, name):
        raise ValueError("name is not valid")


class UDTExtraOptions(BaseModel):
    allocation_cores: Optional[int] = None
    allocation_memory: Optional[int] = None

    sub_type: Optional[str] = None
    target_labels: List[str] = None
    source_column: Optional[str] = None
    target_column: Optional[str] = None
    default_tag: Optional[str] = None
    delimiter: Optional[str] = None
    text_column: Optional[str] = None
    label_column: Optional[str] = None
    n_target_classes: Optional[int] = None

    @root_validator(pre=True)
    def set_fields_based_on_type(cls, values):
        sub_type = values.get("sub_type")
        if sub_type == "text":
            values["delimiter"] = values.get("delimiter", ",")
            values["text_column"] = values.get("text_column", "text")
            values["label_column"] = values.get("label_column", "label")
            values["n_target_classes"] = values.get("n_target_classes", None)
        elif sub_type == "token":
            values["target_labels"] = values.get("target_labels", [])
            values["source_column"] = values.get("source_column", "source")
            values["target_column"] = values.get("target_column", "target")
            values["default_tag"] = values.get("default_tag", "O")
        return values


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
    checkpoint_interval: Optional[int] = None
    fast_approximation: Optional[bool] = None
    num_buckets_to_sample: Optional[int] = None
    metrics: Optional[list[str]] = None
    on_disk: Optional[bool] = None
    docs_on_disk: Optional[bool] = None

    class Config:
        extra = "forbid"


def get_model(session: Session, username: str, model_name: str):
    return (
        session.query(schema.Model)
        .join(schema.User)
        .filter(schema.User.username == username, schema.Model.name == model_name)
        .first()
    )


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


def update_json_list(current_list, new_dict):
    if current_list is None:
        current_list = []
    else:
        current_list = json.loads(current_list)

    current_list.append(new_dict)
    return json.dumps(current_list)


def get_deployment(session: Session, deployment_name, deployment_user_id, model_id):
    return (
        session.query(schema.Deployment)
        .filter(
            schema.Deployment.name == deployment_name,
            schema.Deployment.user_id == deployment_user_id,
            schema.Deployment.model_id == model_id,
        )
        .first()
    )


def model_accessible(model: schema.Model, user: schema.User) -> bool:
    if model.access_level == "private":
        if model.user.id == user.id:
            return True
        return False

    if model.access_level == "protected":
        if model.domain == user.domain:
            return True
        return False

    return True


def get_empty_port():
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.bind(("", 0))  # Bind to an empty
    port = sock.getsockname()[1]
    sock.close()
    return port


def parse_deployment_identifier(deployment_identifier):
    regex_pattern = "^[\w-]+\/[\w-]+\:[\w-]+\/[\w-]+$"
    if re.match(regex_pattern, deployment_identifier):
        model_identifier, deployment_tag = deployment_identifier.split(":")
        model_username, model_name = model_identifier.split("/")
        deployment_username, deployment_name = deployment_tag.split("/")
        return model_username, model_name, deployment_username, deployment_name
    else:
        raise ValueError("deployment identifier is not valid")


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


GENERATE_JOB_ID = "llm-generation"


async def restart_generate_job():
    nomad_endpoint = os.getenv("NOMAD_ENDPOINT")
    if nomad_job_exists(GENERATE_JOB_ID, nomad_endpoint):
        delete_nomad_job(GENERATE_JOB_ID, nomad_endpoint)
    cwd = Path(os.getcwd())
    platform = get_platform()
    return submit_nomad_job(
        nomad_endpoint=nomad_endpoint,
        filepath=str(cwd / "backend" / "nomad_jobs" / "generation_job.hcl.j2"),
        platform=platform,
        port=None if platform == "docker" else get_empty_port(),
        tag=os.getenv("TAG"),
        registry=os.getenv("DOCKER_REGISTRY"),
        docker_username=os.getenv("DOCKER_USERNAME"),
        docker_password=os.getenv("DOCKER_PASSWORD"),
        image_name=os.getenv("GENERATION_IMAGE"),
        python_path=get_python_path(),
        generate_app_dir=str(get_root_absolute_path() / "llm_generation_job"),
    )
