import os
import traceback
import uuid
from enum import Enum
from pathlib import Path
from typing import Dict, List, Optional

import pandas as pd
from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import (
    get_platform,
    get_python_path,
    get_root_absolute_path,
    response,
    submit_nomad_job,
)
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, Form, status
from licensing.verify.verify_license import valid_job_allocation, verify_license
from pydantic import BaseModel, ValidationError
from sqlalchemy.orm import Session

data_router = APIRouter()

COLUMNS = {
    schema.UDT_Task.TEXT: {"source": "text", "target": "label"},
    schema.UDT_Task.TOKEN: {"source": "source", "target": "target"},
}


def update_tags(tag_sequence: str, tags_to_keep: List[str]):
    tags = tag_sequence.split()
    updated_tags = list(map(lambda tag: "O" if tag not in tags_to_keep else tag, tags))
    return " ".join(updated_tags)


def replace_tags_and_write(
    catalog: schema.Catalog, target_labels: List[str], write_path: str
):
    train_path = os.path.join(
        os.getenv("SHARE_DIR"), "datasets", str(catalog.id), "train.csv"
    )
    df = pd.read_csv(train_path)

    # TODO(Gautam/pratyush): Pass this source and target column from platform to container to stay consistent with it
    df[COLUMNS[schema.UDT_Task.TEXT]["target"]] = df[
        COLUMNS[schema.UDT_Task.TEXT]["target"]
    ].apply(update_tags, tags_to_keep=target_labels)
    df.to_csv(write_path, mode="a", index=False, header=not os.path.exists(write_path))
    return len(df)


def prune_labels_and_write(
    catalog: schema.Catalog, target_labels: List[str], write_path: str
):
    most_suited_dataset = os.path.join(
        os.getenv("SHARE_DIR"), "datasets", str(catalog.id), "train.csv"
    )
    df = pd.read_csv(most_suited_dataset)
    df = df[df[COLUMNS[schema.UDT_Task.TEXT]["target"]].isin(target_labels)]
    df.to_csv(write_path, mode="a", index=False, header=not os.path.exists(write_path))
    return len(df)


def prune_and_merge(
    task: schema.UDT_Task,
    existing_datasets: List[schema.Catalog],
    target_labels: List[str],
    data_id: str,
):
    save_dir = os.path.join(os.getenv("SHARE_DIR"), str(data_id))
    os.makedirs(save_dir, exist_ok=True)
    train_file = os.path.join(save_dir, "train.csv")
    samples_generated = 0
    for catalog in existing_datasets:
        if task == schema.UDT_Task.TOKEN:
            samples_generated += replace_tags_and_write(
                catalog, target_labels, train_file
            )
        else:
            samples_generated += prune_labels_and_write(
                catalog, target_labels, train_file
            )
    return samples_generated


def find_catalogs(catalogs: List[schema.Catalog], target_labels: str):
    def similarity(dataset_labels: List[str], target_labels: List[str]) -> float:
        if not dataset_labels:
            return 0.0
        match_count = len(set(dataset_labels) & set(target_labels))
        return match_count / len(target_labels)

    most_suited_datasets = list(
        filter(lambda x: similarity(x.target_labels, target_labels) == 1, catalogs)
    )

    return most_suited_datasets


# Not sure if this will lead to exposure of using openai to generate data
class LLMProvider(str, Enum):
    openai = "openai"
    cohere = "cohere"


class TextClassificationGenerateArgs(BaseModel):
    samples_per_label: int
    target_labels: List[str]
    examples: Dict[str, List[str]]
    labels_description: Dict[str, str]
    user_vocab: Optional[List[str]] = None
    user_prompts: Optional[List[str]] = None
    vocab_per_sentence: int = 4
    allocation_cores: Optional[int] = None
    allocation_memory: Optional[int] = None


@data_router.post("/generate-text-data")
def generate_text_data(
    task_prompt: str,
    llm_provider: LLMProvider = LLMProvider.openai,
    form: str = Form(default="{}"),
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    # TODO(Gautam): Only people from ThirdAI should be able to access this endpoint
    try:
        extra_options = TextClassificationGenerateArgs.model_validate_json(
            form
        ).model_dump()
        extra_options = {k: v for k, v in extra_options.items() if v is not None}
        if extra_options:
            print(f"Extra options for training: {extra_options}")
    except ValidationError as e:
        return {"error": "Invalid extra options format", "details": str(e)}

    try:
        license_info = verify_license(
            os.getenv(
                "LICENSE_PATH", "/model_bazaar/license/ndb_enterprise_license.json"
            )
        )
        if not valid_job_allocation(license_info, os.getenv("NOMAD_ENDPOINT")):
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=f"Resource limit reached, cannot allocate new jobs.",
            )
    except Exception as e:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"License is not valid. {str(e)}",
        )

    data_id = uuid.uuid4()

    genai_key = os.getenv("GENAI_KEY")
    if genai_key is None:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Need gen_ai key for data-generation",
        )

    existing_datasets = find_datasets(
        task=schema.UDT_Task.TEXT, target_labels=extra_options["target_labels"]
    )
    samples_found = 0
    if existing_datasets:
        samples_found = prune_and_merge(
            existing_datasets, extra_options["target_labels"], data_id
        )

    submit_nomad_job(
        str(Path(os.getcwd()) / "backend" / "nomad_jobs" / "generate_data_job.hcl.j2"),
        nomad_endpoint=os.getenv("NOMAD_ENDPOINT"),
        platform=get_platform(),
        tag=os.getenv("TAG"),
        registry=os.getenv("DOCKER_REGISTRY"),
        docker_username=os.getenv("DOCKER_USERNAME"),
        docker_password=os.getenv("DOCKER_PASSWORD"),
        image_name=os.getenv("TRAIN_IMAGE_NAME"),
        train_script=str(get_root_absolute_path() / "data_generation/run.py"),
        task_prompt=task_prompt,
        data_id=str(data_id),
        data_category="text",
        llm_provider=llm_provider.value,
        model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
        share_dir=os.getenv("SHARE_DIR", None),
        genai_key=os.getenv("GENAI_KEY", None),
        license_key=license_info["boltLicenseKey"],
        extra_options=extra_options,
        python_path=get_python_path(),
        sentences_generated=samples_found,
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully submitted the data-generation job",
    )


class TokenClassificationGenerateArgs(BaseModel):
    domain_prompt: str
    tags: List[str]
    tag_examples: Dict[str, List[str]]
    num_sentences_to_generate: int
    num_samples_per_tag: int = 4
    allocation_cores: Optional[int] = None
    allocation_memory: Optional[int] = None


@data_router.post("/generate-token-data")
def generate_token_data(
    task_prompt: str,
    llm_provider: LLMProvider = LLMProvider.openai,
    form: str = Form(default="{}"),
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    # TODO(Gautam): Only people from ThirdAI should be able to access this endpoint
    try:
        extra_options = TokenClassificationGenerateArgs.model_validate_json(
            form
        ).model_dump()
        extra_options = {k: v for k, v in extra_options.items() if v is not None}
        if extra_options:
            print(f"Extra options for training: {extra_options}")
    except ValidationError as e:
        return {"error": "Invalid extra options format", "details": str(e)}

    try:
        license_info = verify_license(
            os.getenv(
                "LICENSE_PATH", "/model_bazaar/license/ndb_enterprise_license.json"
            )
        )
        if not valid_job_allocation(license_info, os.getenv("NOMAD_ENDPOINT")):
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=f"Resource limit reached, cannot allocate new jobs.",
            )
    except Exception as e:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"License is not valid. {str(e)}",
        )

    data_id = uuid.uuid4()

    genai_key = os.getenv("GENAI_KEY")
    if genai_key is None:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Need gen_ai key for data-generation",
        )

    existing_datasets = find_datasets(
        task=schema.UDT_Task.TOKEN, target_labels=extra_options["tags"]
    )
    samples_found = 0
    if existing_datasets:
        samples_found = prune_and_merge(
            existing_datasets, extra_options["tags"], data_id
        )

    submit_nomad_job(
        str(Path(os.getcwd()) / "backend" / "nomad_jobs" / "generate_data_job.hcl.j2"),
        nomad_endpoint=os.getenv("NOMAD_ENDPOINT"),
        platform=get_platform(),
        tag=os.getenv("TAG"),
        registry=os.getenv("DOCKER_REGISTRY"),
        docker_username=os.getenv("DOCKER_USERNAME"),
        docker_password=os.getenv("DOCKER_PASSWORD"),
        image_name=os.getenv("TRAIN_IMAGE_NAME"),
        train_script=str(get_root_absolute_path() / "data_generation/run.py"),
        task_prompt=task_prompt,
        data_id=str(data_id),
        data_category="token",
        llm_provider=llm_provider.value,
        model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
        share_dir=os.getenv("SHARE_DIR", None),
        genai_key=genai_key,
        license_key=license_info["boltLicenseKey"],
        extra_options=extra_options,
        python_path=get_python_path(),
        sentences_generated=samples_found,
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully submitted the data-generation job",
    )


# @data_router.post("/find-dataset")
def find_datasets(
    task: schema.UDT_Task,
    target_labels: List[str],
    session: Session = Depends(get_session),
):

    try:
        catalogs: List[schema.Catalog] = (
            session.query(schema.Catalog).filter(schema.Catalog.task == task).all()
        )
        # Filtering catalogs based on the target_labels
        most_suited_dataset_catalogs = find_catalogs(
            catalogs, target_labels=target_labels
        )
        return most_suited_dataset_catalogs

    except Exception as e:
        traceback.print_exc()
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="unable to find a sample dataset",
        )


class GenerationComplete(BaseModel):
    data_id: str
    task: schema.UDT_Task
    target_labels: List[str]
    samples_generated: int


@data_router.post()
def generation_complete(
    body: GenerationComplete, session: Session = Depends(get_session)
):
    """
    Mark the training of a model as complete.

    Parameters:
    - body: The body of the request containing data_id.
        - Example:
        ```json
        {
            "data_id": "123e4567-e89b-12d3-a456-426614174000",
            "samples_generated": 100
        }
        ```
    - session: The database session (dependency).

    Returns:
    - A JSON response indicating the update status.
    """

    catalog_entry = schema.Catalog(
        data_id=body.data_id,
        name="Generated_data",
        task=body.task,
        target_labels=body.target_labels,
    )

    session.add(catalog_entry)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully updated"
    )
