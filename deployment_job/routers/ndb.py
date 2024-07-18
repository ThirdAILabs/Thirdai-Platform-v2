import logging
import traceback
import uuid
from multiprocessing import Lock, Manager, Queue
from typing import List, Optional

import thirdai
from fastapi import APIRouter, Depends, Form, UploadFile, status
from fastapi.encoders import jsonable_encoder
from permissions import Permissions
from pydantic import ValidationError
from pydantic_models import inputs
from pydantic_models.documents import DocumentList
from pydantic_models.inputs import BaseQueryParams, NDBExtraParams
from routers.model import get_model
from utils import (
    Status,
    log_function_name,
    logger,
    now,
    propagate_error,
    response,
    validate_files,
    validate_name,
)
from variables import GeneralVariables, TypeEnum

ndb_router = APIRouter()
permissions = Permissions()

general_variables = GeneralVariables.load_from_env()


@log_function_name
@ndb_router.post("/predict")
@propagate_error
def ndb_query(
    base_params: BaseQueryParams,
    ndb_params: Optional[NDBExtraParams] = NDBExtraParams(),
    _=Depends(permissions.verify_read_permission),
):
    model = get_model()
    params = base_params.model_dump()
    if general_variables.type == TypeEnum.NDB:
        extra_params = ndb_params.model_dump(exclude_unset=True)
        params.update(extra_params)

    results = model.predict(**params)

    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=jsonable_encoder(results),
    )


@log_function_name
@ndb_router.post("/upvote")
@propagate_error
def ndb_upvote(
    input: inputs.UpvoteInput, token=Depends(permissions.verify_write_permission)
):
    model = get_model()
    model.upvote(text_id_pairs=input.text_id_pairs, token=token)

    return response(status_code=status.HTTP_200_OK, message="Sucessfully upvoted")


@log_function_name
@ndb_router.post("/associate")
@propagate_error
def ndb_associate(
    input: inputs.AssociateInput,
    token=Depends(permissions.verify_write_permission),
):
    model = get_model()
    model.associate(text_pairs=input.text_pairs, token=token)

    return response(status_code=status.HTTP_200_OK, message="Sucessfully associated")


@log_function_name
@ndb_router.get("/sources")
@propagate_error
def get_sources(_=Depends(permissions.verify_read_permission)):
    model = get_model()
    sources = model.sources()
    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=sources,
    )


@log_function_name
@ndb_router.post("/delete")
@propagate_error
def delete(input: inputs.DeleteInput, _=Depends(permissions.verify_write_permission)):
    model = get_model()
    model.delete(source_ids=input.source_ids)

    return response(
        status_code=status.HTTP_200_OK,
        message=f"{len(input.source_ids)} file(s) deleted",
        success=True,
    )


@log_function_name
@ndb_router.post("/save")
def save(
    input: inputs.SaveModel,
    token=Depends(permissions.verify_read_permission),
    override_permission=Depends(permissions.get_owner_permission),
):
    model = get_model()
    model_id = general_variables.model_id
    if not input.override:
        model_id = str(uuid.uuid4())
        if not input.model_name:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Model name is required for new model.",
            )

        try:
            validate_name(input.model_name)
        except Exception:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Name must only contain alphanumeric characters, underscores (_), and hyphens (-). ",
            )

        is_model_present = model.reporter.check_model_present(token, input.model_name)
        if is_model_present:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Model name already exists, choose another one.",
            )
    else:
        if not override_permission:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="You dont have permissions to override this model.",
            )
    try:
        model.save_ndb(model_id=model_id)
        if not input.override:
            model.reporter.save_model(
                access_token=token,
                deployment_id=general_variables.deployment_id,
                model_id=model_id,
                base_model_id=general_variables.model_id,
                model_name=input.model_name,
                metadata={"thirdai_version": str(thirdai.__version__)},
            )
    except Exception as err:
        traceback.print_exc()
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, message=str(err)
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully saved the model.",
        data={"new_model_id": model_id if not input.override else None},
    )


task_queue = Queue()
tasks = {}
task_lock = Lock()


@log_function_name
@ndb_router.post("/insert")
@propagate_error
def insert(
    documents: str = Form(...),
    files: List[UploadFile] = [],
    input_mode: str = Form("sync"),
    _=Depends(permissions.verify_write_permission),
):
    try:
        documents_list = DocumentList.model_validate_json(documents).model_dump()
    except ValidationError as e:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="Invalid format for document insertion.",
            data={"details": str(e), "documents": documents},
        )

    if not documents_list:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="No documents supplied for insertion. Must supply at least one document.",
        )

    model = get_model()

    validate_files(documents_list, files, model.data_dir)

    if input_mode == "async":
        task_id = str(uuid.uuid4())
        with task_lock:
            tasks[task_id] = {
                "status": Status.not_started,
                "action": "insert",
                "last_modified": now(),
                "documents": documents_list,
                "message": "",
                "data": None,
            }
        task_queue.put((task_id, documents_list))
        return response(
            status_code=status.HTTP_202_ACCEPTED,
            message="Files received, insertion will start soon.",
            data={"task_id": task_id},
        )

    try:
        sources = model.insert(documents=documents_list)
    except Exception as err:
        logging.error(f"Failed to insert documents: {err}")
        traceback.print_exc()
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="Failed to insert the files",
            success=False,
        )

    return response(
        status_code=status.HTTP_200_OK,
        message=f"{len(sources)} documents uploaded to ndb model",
        data=sources,
        success=True,
    )


@log_function_name
@ndb_router.post("/task-status")
@propagate_error
def task_status(
    task_id: str,
    _=Depends(permissions.verify_write_permission),
):
    with task_lock:
        if task_id in tasks:
            return response(
                status_code=status.HTTP_200_OK,
                message=f"Information for task {task_id}",
                data=tasks[task_id],
            )
        else:
            return response(
                status_code=status.HTTP_404_NOT_FOUND,
                message="Task ID not found",
            )


def process_tasks():
    model = get_model()
    while True:
        task_id, documents = task_queue.get()
        try:
            with task_lock:
                tasks[task_id]["status"] = Status.in_progress
                tasks[task_id]["last_modified"] = now()

            sources = model.insert(documents=documents)

            with task_lock:
                tasks[task_id]["status"] = Status.complete
                tasks[task_id]["data"] = sources
                tasks[task_id]["last_modified"] = now()

        except Exception as e:
            with task_lock:
                tasks[task_id]["status"] = Status.failed
                tasks[task_id]["message"] = str(traceback.format_exc())
                tasks[task_id]["last_modified"] = now()
                logging.error(
                    f"Task {task_id} with data {tasks[task_id]} failed: {str(traceback.format_exc())}"
                )
