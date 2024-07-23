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
    UDTExtraOptions,
    get_files,
    get_model,
    get_model_from_identifier,
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
from fastapi.encoders import jsonable_encoder
from licensing.verify.verify_license import valid_job_allocation, verify_license
from pydantic import BaseModel, ValidationError
from sqlalchemy.orm import Session

train_router = APIRouter()


@train_router.post("/ndb")
def train(
    model_name: str,
    files: List[UploadFile],
    file_details_list: Optional[str] = Form(default=None),
    base_model_identifier: Optional[str] = None,
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

    if base_model_identifier:
        try:
            base_model: schema.Model = get_model_from_identifier(
                base_model_identifier, session
            )
        except Exception as error:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=str(error),
            )

    try:
        new_model: schema.Model = schema.Model(
            id=model_id,
            user_id=user.id,
            train_status=schema.Status.not_started,
            name=model_name,
            type="ndb",
            sub_type=None,
            domain=user.email.split("@")[1],
            access_level=schema.Access.private,
            parent_id=base_model.id if base_model_identifier else None,
        )

        session.add(new_model)
        session.commit()
        session.refresh(new_model)

        work_dir = os.getcwd()

        sharded = (
            True
            if extra_options.get("num_models_per_shard") > 1
            or extra_options.get("num_shards") > 1
            else False
        )

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
            model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
            share_dir=os.getenv("SHARE_DIR", None),
            license_key=license_info["boltLicenseKey"],
            extra_options=extra_options,
            python_path=get_python_path(),
            aws_access_key=(os.getenv("AWS_ACCESS_KEY", "")),
            aws_access_secret=(os.getenv("AWS_ACCESS_SECRET", "")),
            base_model_id=("NONE" if not base_model_identifier else str(base_model.id)),
            type="ndb",
            sub_type="normal" if not sharded else "shard_allocation",
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


@train_router.post("/udt")
def train(
    model_name: str,
    files: List[UploadFile],
    file_details_list: Optional[str] = Form(default=None),
    base_model_identifier: Optional[str] = None,
    extra_options_form: str = Form(default="{}"),
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user
    try:
        extra_options = UDTExtraOptions.parse_raw(extra_options_form).dict()
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

    if base_model_identifier:
        try:
            base_model: schema.Model = get_model_from_identifier(
                base_model_identifier, session
            )
        except Exception as error:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=str(error),
            )

    try:
        new_model: schema.Model = schema.Model(
            id=model_id,
            user_id=user.id,
            train_status=schema.Status.not_started,
            name=model_name,
            type="udt",
            sub_type=extra_options['sub_type'],
            domain=user.email.split("@")[1],
            access_level=schema.Access.private,
            parent_id=base_model.id if base_model_identifier else None,
        )

        session.add(new_model)
        session.commit()
        session.refresh(new_model)

        work_dir = os.getcwd()

        udt_subtype = extra_options['sub_type']
        extra_options.pop('sub_type', None)
        
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
            model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
            share_dir=os.getenv("SHARE_DIR", None),
            license_key=license_info["boltLicenseKey"],
            extra_options=extra_options,
            python_path=get_python_path(),
            aws_access_key=(os.getenv("AWS_ACCESS_KEY", "")),
            aws_access_secret=(os.getenv("AWS_ACCESS_SECRET", "")),
            base_model_id=("NONE" if not base_model_identifier else str(base_model.id)),
            type="udt",
            sub_type=udt_subtype,
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
        metadata.train = update_json(metadata.train, body.metadata)
    else:
        new_metadata = schema.MetaData(
            model_id=trained_model.id,
            train=body.metadata,
        )
        session.add(new_metadata)

    session.commit()

    return {"message": "successfully updated"}


@train_router.post("/update-status")
def train_fail(
    model_id: str,
    train_status: schema.Status,
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

    trained_model.train_status = train_status
    session.commit()

    return {"message": f"successfully updated with following {message}"}


@train_router.post("/create-shard")
def create_shard(
    shard_num: int,
    model_id: str,
    data_id: str,
    base_model_id: Optional[str] = None,
    extra_options_form: str = Form(default="{}"),
    session: Session = Depends(get_session),
):
    try:
        extra_options = NDBExtraOptions.parse_raw(extra_options_form).dict()
        extra_options = {k: v for k, v in extra_options.items() if v is not None}
        if extra_options:
            print(f"Extra options for shard training: {extra_options}")
    except ValidationError as e:
        return {"error": "Invalid extra options format", "details": str(e)}

    try:
        license_info = verify_license(
            os.getenv(
                "LICENSE_PATH", "/model_bazaar/license/ndb_enterprise_license.json"
            )
        )
    except Exception as e:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"License is not valid. {str(e)}",
        )

    try:
        new_shard: schema.ModelShard = schema.ModelShard(
            model_id=model_id,
            shard_num=shard_num,
            train_status=schema.Status.not_started,
        )
        session.add(new_shard)
        session.commit()

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
            model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT", None),
            share_dir=os.getenv("SHARE_DIR", None),
            license_key=license_info["boltLicenseKey"],
            extra_options=extra_options,
            python_path=get_python_path(),
            type="ndb",
            sub_type="shard_train",
            base_model_id="NONE" if not base_model_id else base_model_id,
            shard_num=shard_num,
        )

    except Exception as err:
        logger.info(str(err))
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message=str(err),
        )

    return {"message": f"Successfully created shard"}


@train_router.post("/update-shard-train-status")
def update_shard_train_status(
    shard_num: int,
    model_id: str,
    status: schema.Status,
    message: str = "",
    session: Session = Depends(get_session),
):
    model_shard: schema.ModelShard = (
        session.query(schema.ModelShard)
        .filter(
            schema.ModelShard.model_id == model_id,
            schema.ModelShard.shard_num == shard_num,
        )
        .first()
    )

    if not model_shard:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"No model shard with id {model_id} and shard number {shard_num}.",
        )

    model_shard.train_status = status
    session.commit()

    return {"message": f"Successfully updated shard with message: {message}"}


@train_router.get("/status")
def train_status(
    model_identifier: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    Get the status of a NeuralDB.

    - **model_identifier**: model identifier of model to retrieve info about

    Returns model status.
    """
    try:
        model: schema.Model = get_model_from_identifier(model_identifier, session)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the train status.",
        data={
            "model_identifier": model_identifier,
            "status": model.train_status,
        },
    )


@train_router.get("/model-shard-train-status")
def model_shard_train_status(
    model_id: str,
    session: Session = Depends(get_session),
):
    try:
        model_shards: List[schema.ModelShard] = (
            session.query(schema.ModelShard)
            .filter(schema.ModelShard.model_id == model_id)
            .all()
        )
    except Exception as error:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message=str(error),
        )

    if not model_shards:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="No shards found for the given model id.",
        )

    results = [
        {
            "shard_num": result.shard_num,
            "status": result.train_status,
        }
        for result in model_shards
    ]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the train status.",
        data=jsonable_encoder(results),
    )
