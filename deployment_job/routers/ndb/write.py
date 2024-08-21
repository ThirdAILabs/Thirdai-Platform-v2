import logging
import traceback
import uuid
from queue import Queue
from threading import Lock

import thirdai
from fastapi import APIRouter, Depends, Form, UploadFile, status
from file_handler import validate_files
from permissions import Permissions
from pydantic import ValidationError
from pydantic_models import inputs
from pydantic_models.documents import DocumentList
from routers.model import get_model
from utils import Status, now, propagate_error, response, validate_name
from variables import GeneralVariables

permissions = Permissions()
general_variables = GeneralVariables.load_from_env()

# Initialize the APIRouter
ndb_write_router = APIRouter()

# Queues and Locks for managing background tasks
task_queue = Queue()
tasks = {}
task_lock = Lock()


@ndb_write_router.post("/insert")
@propagate_error
def insert(
    documents: str = Form(...),
    files: list[UploadFile] = [],
    input_mode: str = Form("sync"),
    token: str = Depends(permissions.verify_write_permission),
):
    """
    Endpoint to insert documents into the NDB model.

    Parameters:
    - documents: str - JSON string containing document data.
    - files: list[UploadFile] - Optional list of files to upload.
    - input_mode: str - Mode of insertion ("sync" or "async").
    - token: str - Authorization token.

    Returns:
    - JSONResponse: Success or failure message.
    """
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

    model = get_model(write_mode=True)
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
                "token": token,
            }
        task_queue.put((task_id, documents_list))
        return response(
            status_code=status.HTTP_202_ACCEPTED,
            message="Files received, insertion will start soon.",
            data={"task_id": task_id},
        )

    try:
        sources = model.insert(documents=documents_list, token=token)
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


@ndb_write_router.post("/associate")
@propagate_error
def ndb_associate(
    input: inputs.AssociateInput,
    token: str = Depends(permissions.verify_read_permission),
):
    """
    Endpoint to associate text pairs in the NDB model.

    Parameters:
    - input: AssociateInput - The input data containing text pairs.
    - token: str - Authorization token.

    Returns:
    - JSONResponse: Success or failure message.
    """
    model = get_model(write_mode=True)
    model.associate(text_pairs=input.text_pairs, token=token)

    return response(status_code=status.HTTP_200_OK, message="Successfully associated")


@ndb_write_router.post("/upvote")
@propagate_error
def ndb_upvote(
    input: inputs.UpvoteInput,
    token: str = Depends(permissions.verify_read_permission),
):
    """
    Endpoint to upvote specific text-id pairs in the NDB model.

    Parameters:
    - input: UpvoteInput - The input data containing text-id pairs.
    - token: str - Authorization token.

    Returns:
    - JSONResponse: Success or failure message.
    """
    model = get_model(write_mode=True)
    model.upvote(text_id_pairs=input.text_id_pairs, token=token)

    return response(status_code=status.HTTP_200_OK, message="Successfully upvoted")


@ndb_write_router.post("/delete")
@propagate_error
def delete(
    input: inputs.DeleteInput,
    token: str = Depends(permissions.verify_write_permission),
):
    """
    Endpoint to delete sources from the model.

    Parameters:
    - input: DeleteInput - The input containing source IDs to be deleted.
    - token: str - Authorization token.

    Returns:
    - JSONResponse: Deletion success message.
    """
    model = get_model(write_mode=True)
    model.delete(source_ids=input.source_ids, token=token)

    return response(
        status_code=status.HTTP_200_OK,
        message=f"{len(input.source_ids)} file(s) deleted",
        success=True,
    )


@ndb_write_router.post("/save")
def save(
    input: inputs.SaveModel,
    token: str = Depends(permissions.verify_read_permission),
    override_permission: bool = Depends(permissions.get_owner_permission),
):
    """
    Endpoint to save the current state of the NDB model.

    Parameters:
    - input: SaveModel - The input parameters for saving the model.
    - token: str - Authorization token.
    - override_permission: bool - Whether the user has permission to override the model.

    Returns:
    - JSONResponse: Save success message.
    """
    model = get_model(write_mode=True)
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
                message="You don't have permissions to override this model.",
            )
    try:
        model.save_ndb(model_id=model_id)
        if not input.override:
            model.reporter.save_model(
                access_token=token,
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


@ndb_write_router.post("/task-status")
@propagate_error
def task_status(
    task_id: str,
    _=Depends(permissions.verify_write_permission),
):
    """
    Endpoint to get the status of a specific task.

    Parameters:
    - task_id: str - The ID of the task.

    Returns:
    - JSONResponse: Task status information.
    """
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
    """
    Processes tasks from the task queue.

    Args:
        task_queue (Queue): Queue for handling background tasks.
        task_lock (Lock): Lock for synchronizing task access.
        tasks (dict): Dictionary to store task details.
    """
    model = get_model(write_mode=True)
    while True:
        task_id, documents = task_queue.get()
        try:
            with task_lock:
                tasks[task_id]["status"] = Status.in_progress
                tasks[task_id]["last_modified"] = now()

            sources = model.insert(documents=documents, token=tasks[task_id]["token"])

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
