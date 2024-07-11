import os
import uuid
from pathlib import Path
from typing import Dict, List, Optional

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import (
    FileDetails,
    FileDetailsList,
    FileType,
    NDBExtraOptions,
    get_files,
    get_model,
    get_platform,
    get_python_path,
    get_root_absolute_path,
    logger,
    response,
    submit_nomad_job,
    update_json,
    validate_name,
)
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, Form, UploadFile, status
from licensing.verify.verify_license import valid_job_allocation, verify_license
from pydantic import BaseModel, ValidationError
from sqlalchemy.orm import Session

train_router = APIRouter()


@train_router.post("/ndb")
def train(
    model_name: str,
    files: List[UploadFile],
    file_details_list: Optional[str] = Form(default=None),
    extra_options_form: str = Form(default="{}"),
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user
    try:
        extra_options = NDBExtraOptions.parse_raw(extra_options_form).dict()
        extra_options = {k: v for k, v in extra_options.items() if v is not None}
        if extra_options:
            print(f"Extra options for training: {extra_options}")
    except ValidationError as e:
        return {"error": "Invalid extra options format", "details": str(e)}

    if file_details_list:
        try:
            files_info = FileDetailsList.parse_raw(file_details_list).file_details
        except ValidationError as e:
            return {"error": "Invalid file details list format", "details": str(e)}
    else:
        files_info = [
            FileDetails(mode=FileType.unsupervised, location="local") for _ in files
        ]

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

    try:
        validate_name(model_name)
    except:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"{model_name} is not a valid model name.",
        )

    duplicate_model = get_model(session, username=user.username, model_name=model_name)
    if duplicate_model:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Model with name {model_name} already exists for user {user.username}.",
        )

    model_id = uuid.uuid4()
    data_id = model_id

    if len(files) != len(files_info):
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Given {len(files)} files but for {len(files_info)} files the info has given.",
        )

    filenames = get_files(files, data_id, files_info)

    if not isinstance(filenames, list):
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=filenames,
        )

    if len(filenames) == 0:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"No files provided.",
        )

    unique_filenames = set(filenames)
    if len(filenames) != len(unique_filenames):
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Duplicate filenames recieved, please ensure each filename is unique.",
        )

    try:
        new_model: schema.Model = schema.Model(
            id=model_id,
            user_id=user.id,
            train_status=schema.Status.not_started,
            name=model_name,
            type="ndb",
            domain=user.email.split("@")[1],
            access_level=schema.Access.private,
        )

        session.add(new_model)
        session.commit()
        session.refresh(new_model)

        work_dir = os.getcwd()

        submit_nomad_job(
            str(Path(work_dir) / "backend" / "nomad_jobs" / "train_job.hcl.j2"),
            nomad_endpoint=os.getenv("NOMAD_ENDPOINT"),
            platform=get_platform(),
            tag=os.getenv("TAG"),
            registry=os.getenv("DOCKER_REGISTRY"),
            docker_username=os.getenv("DOCKER_USERNAME"),
            docker_password=os.getenv("DOCKER_PASSWORD"),
            image_name=os.getenv("TRAIN_IMAGE_NAME"),
            train_script=str(get_root_absolute_path() / "train_job/run.py"),
            model_id=str(model_id),
            data_id=str(data_id),
            user_id=str(user.id),
            model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
            share_dir=os.getenv("SHARE_DIR", None),
            license_key=license_info["boltLicenseKey"],
            extra_options=extra_options,
            python_path=get_python_path(),
            aws_access_key=(os.getenv("AWS_ACCESS_KEY", "")),
            aws_access_secret=(os.getenv("AWS_ACCESS_SECRET", "")),
        )

    except Exception as err:
        # TODO: change the status of the new model entry to failed

        logger.info(str(err))
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message=str(err),
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully submitted the job",
        data={
            "model_id": str(model_id),
            "user_id": str(user.id),
        },
    )


class TrainComplete(BaseModel):
    model_id: str
    metadata: Dict[str, str]


@train_router.post("/complete")
def train_complete(
    body: TrainComplete,
    session: Session = Depends(get_session),
):
    trained_model: schema.Model = (
        session.query(schema.Model).filter(schema.Model.id == body.model_id).first()
    )

    if not trained_model:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"No model with id {body.model_id}.",
        )

    trained_model.train_status = schema.Status.complete
    trained_model.access_level = schema.Access.private

    metadata: schema.MetaData = trained_model.meta_data
    if metadata:
        metadata.public = update_json(metadata.public, body.metadata)
    else:
        new_metadata = schema.MetaData(
            model_id=trained_model.id,
            public=body.metadata,
        )
        session.add(new_metadata)

    session.commit()

    return {"message": "successfully updated"}


@train_router.post("/update-status")
def train_fail(
    model_id: str,
    status: schema.Status,
    message: str,
    session: Session = Depends(get_session),
):
    trained_model: schema.Model = (
        session.query(schema.Model).filter(schema.Model.id == model_id).first()
    )

    if not trained_model:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"No model with id {model_id}.",
        )

    trained_model.train_status = status
    session.commit()

    return {"message": f"successfully updated with following {message}"}
