import datetime
import os
import traceback
import uuid
from pathlib import Path
from typing import Dict, List, Optional

import pytz

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import (
    get_platform,
    get_python_path,
    get_root_absolute_path,
    response,
    submit_nomad_job,
    update_task_status,
    State,
    get_task,
    filter_by,
    find_dataset,
)
from database.session import get_session
from fastapi import APIRouter, Depends, Form, status
from licensing.verify.verify_license import valid_job_allocation, verify_license
from pydantic import BaseModel, ValidationError
from sqlalchemy.orm import Session
from database import schema

data_router = APIRouter()

class TextClassificationGenerateArgs(BaseModel):
    samples_per_label: int
    target_labels: List[str]
    examples: Dict[str, List[str]]
    labels_description: Dict[str, str]
    user_vocab: Optional[List[str]] = None
    user_prompts: Optional[List[str]] = None
    batch_size: int = 40
    vocab_per_sentence: int = 4


class TokenClassificationGenerateArgs(BaseModel):
    domain_prompt: str
    tags: List[str]
    tag_examples: Dict[str, List[str]]
    num_call_batches: int
    batch_size: int = 40
    num_samples_per_tag: int = 4
    
@data_router.post("/find-text-dataset")
def find_text_datasets(
    workflow_id: uuid.UUID,
    args: TextClassificationGenerateArgs,
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    # Infer task should be completed by now
    with get_session() as session:
        workflow = session.query(schema.Workflow).filter(schema.Workflow.id == workflow_id).first()

    if workflow.infer_task_status != State.DONE:
        return response(
            status_code=status.HTTP_405_METHOD_NOT_ALLOWED,
            message="Infer task is not completed yet",
        )
    elif workflow.datasource_status == State.DONE:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="Data is already generated or found",
        )

    try:
        # check if datasource exists
        with get_session() as session:
            existing_datasource: schema.DataSource = (
                session.query(schema.DataSource)
                .filter(schema.DataSource.workflow_id == workflow_id)
                .first()
            )
        
        if existing_datasource:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=f"status: {workflow.datasource_status}",
            )

        datasource_id = uuid.uuid4()

        with get_session() as session:
            entry = schema.DataSource(
                id=datasource_id,
                workflow_id=workflow_id,
                samples_per_label=args.samples_per_label,
                target_labels=args.target_labels,
                user_vocab=args.user_vocab,
                examples=args.examples,
                user_prompts=args.user_prompts,
                labels_description=args.labels_description,
                batch_size=args.batch_size,
                vocab_per_sentence=args.vocab_per_sentence,
                user_inserted=False,
                generated=True,
            )
            session.add(entry)

            # Updating the datasource task status
            workflow_entry = (
                session.query(schema.Workflow).filter(schema.Workflow.id == workflow_id).first()
            )
            workflow_entry.datasource_status = State.RUNNING
            session.commit()
            
        inferred_task = get_task(workflow_id=workflow_id)
        
        filtered_catalogs = filter_by(
            task=inferred_task.task, sub_tasks=inferred_task.sub_tasks
        )

        # Filtering catalogs based on the target_labels
        most_suited_dataset_catalog = None
        if filtered_catalogs:
            most_suited_dataset_catalog = find_dataset(
                catalogs=filtered_catalogs, target_labels=args.target_labels
            )

            if most_suited_dataset_catalog:
                with schema.get_session() as session:
                    entry = (
                        session.query(schema.DataSource)
                        .filter(schema.DataSource.id == datasource_id)
                        .first()
                    )
                    entry.catalog_id = most_suited_dataset_catalog.id
                    entry.generated = False
                    session.commit()
                
                data = {
                    "dataset_name": most_suited_dataset_catalog.name,
                    "catalog_id": str(most_suited_dataset_catalog.id),
                    "find_status": True,
                }

            # print("Found dataset", most_suited_dataset_catalog)

        if not most_suited_dataset_catalog:
            data = {
                    "dataset_name": None,
                    "catalog_id": None,
                    "find_status": False,
                }

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully retrieved the preview of the data",
            data=data,
        )
        
    except Exception as e:
        traceback.print_exc()
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="unable to find a sample text-dataset",
        )

@data_router.post("/find-token-dataset")
def find_token_datasets(
    workflow_id: uuid.UUID,
    args: TokenClassificationGenerateArgs,
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    # Infer task should be completed by now
    with get_session() as session:
        workflow = session.query(schema.Workflow).filter(schema.Workflow.id == workflow_id).first()

    if workflow.infer_task_status != State.DONE:
        return response(
            status_code=status.HTTP_405_METHOD_NOT_ALLOWED,
            message="Infer task is not completed yet",
        )
    elif workflow.datasource_status == State.DONE:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="Data is already generated or found",
        )
    try:
        # check if datasource exists
        with get_session() as session:
            existing_datasource: schema.DataSource = (
                session.query(schema.DataSource)
                .filter(schema.DataSource.workflow_id == workflow_id)
                .first()
            )
        if existing_datasource:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=f"status: {workflow.datasource_status}",
            )
        datasource_id = uuid.uuid4()

        with get_session() as session:
            entry = schema.DataSource(
                id=datasource_id,
                workflow_id=workflow_id,
                domain_prompt=args.domain_prompt,
                tags=args.tags,
                tag_examples=args.tag_examples,
                num_call_batches=args.num_call_batches,
                batch_size=args.batch_size,
                num_samples_per_tag=args.num_samples_per_tag,
                user_inserted=False,
                generated=True,
            )
            session.add(entry)

            # Updating the datasource task status
            workflow_entry = (
                session.query(schema.Workflow).filter(schema.Workflow.id == workflow_id).first()
            )
            workflow_entry.datasource_status = State.RUNNING
            session.commit()


        inferred_task = get_task(workflow_id=workflow_id)
        filtered_catalogs = filter_by(task=inferred_task.task)

        most_suited_dataset_catalog = None
        if filtered_catalogs:
            most_suited_dataset_catalog = find_dataset(
                catalogs=filtered_catalogs, target_labels=args.tags
            )
            if most_suited_dataset_catalog:
                with schema.get_session() as session:
                    entry = (
                        session.query(schema.DataSource)
                        .filter(schema.DataSource.id == datasource_id)
                        .first()
                    )
                    entry.catalog_id = most_suited_dataset_catalog.id
                    entry.generated = False
                    session.commit()
                    
            data = {
                    "dataset_name": most_suited_dataset_catalog.name,
                    "catalog_id": str(most_suited_dataset_catalog.id),
                    "find_status": True,
                }

        if not most_suited_dataset_catalog:
            # No dataset found, Generate dataset
            data = {
                    "dataset_name": None,
                    "catalog_id": None,
                    "find_status": False,
                }
            
        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully retrieved the preview of the data",
            data=data,
        )

    except Exception as e:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="unable to find a sample token-dataset",
        )

@data_router.post("/generate-text-data")
def generate_text_data(
    task_prompt: str,
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
        model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
        share_dir=os.getenv("SHARE_DIR", None),
        genai_key=os.getenv("GENAI_KEY", None),
        license_key=license_info["boltLicenseKey"],
        extra_options=extra_options,
        python_path=get_python_path(),
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully submitted the data-generation job",
    )

@data_router.post("/generate-token-data")
def generate_token_data(
    task_prompt: str,
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
        model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
        share_dir=os.getenv("SHARE_DIR", None),
        genai_key=genai_key,
        license_key=license_info["boltLicenseKey"],
        extra_options=extra_options,
        python_path=get_python_path(),
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully submitted the data-generation job",
    )
