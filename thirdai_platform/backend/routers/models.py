import os
import uuid
from typing import Annotated, Dict, List, Optional, Union

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.routers.utils import (
    get_expiry_min,
    get_high_level_model_info,
    get_model,
    get_model_from_identifier,
    model_accessible,
    response,
    validate_name,
)
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, Header, Query, UploadFile, status, HTTPException
from fastapi.encoders import jsonable_encoder
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
from sqlalchemy import and_, or_
from sqlalchemy.orm import Session, joinedload
from storage import interface, local

from backend.auth_dependencies import team_admin_or_global_admin

model_router = APIRouter()

storage: interface.StorageInterface = local.LocalStorage(
    os.getenv("LOCAL_TEST_DIR", "/model_bazaar")
)


@model_router.get("/public-list")
def list_public_models(
    name: str,
    domain: Optional[str] = None,
    username: Optional[str] = None,
    session: Session = Depends(get_session),
):
    """
    List public models.

    Parameters:
    - name: The name to filter models.
    - domain: Optional domain to filter models.
    - username: Optional username to filter models.
    - session: The database session (dependency).

    Returns:
    - A JSON response with the list of public models.
    """
    results = (
        session.query(schema.Model)
        .options(joinedload(schema.Model.user))
        .filter(
            schema.Model.name.ilike(f"%{name}%"),
            schema.Model.access_level == schema.Access.public,
        )
    )

    results = results.filter(schema.Model.train_status == schema.Status.complete)

    if domain:
        results = results.filter(schema.Model.domain == domain)

    if username:
        results = results.filter(schema.Model.user.username == username)

    results = [get_high_level_model_info(result) for result in results]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the public list",
        data=jsonable_encoder(results),
    )


@model_router.get("/list")
def list_models(
    name: str,
    domain: Optional[str] = None,
    username: Optional[str] = None,
    access_level: Annotated[Union[list[str], None], Query()] = None,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    List models based on the given name, domain, username, and access level.

    Parameters:
    - name: The name to filter models.
    - domain: Optional domain to filter models.
    - username: Optional username to filter models.
    - access_level: Optional access level to filter models.
    - session: The database session (dependency).
    - authenticated_user: The authenticated user (dependency).

    Returns:
    - A JSON response with the list of models.
    """
    user: schema.User = authenticated_user.user

    if user.role == schema.Role.global_admin:
        results = (
            session.query(schema.Model).options(joinedload(schema.Model.user)).all()
        )
    else:
        results = (
            session.query(schema.Model)
            .options(joinedload(schema.Model.user))
            .filter(
                schema.Model.name.ilike(f"%{name}%"),
                or_(
                    # public
                    schema.Model.access_level == schema.Access.public,
                    # protected and matching domain
                    and_(
                        schema.Model.access_level == schema.Access.protected,
                        schema.Model.team_id == user.team_id,
                    ),
                    # private and matching user or admin
                    and_(
                        schema.Model.access_level == schema.Access.private,
                        or_(schema.Model.user_id == user.id, schema.User.id == user.id),
                        or_(user.role == schema.Role.team_admin),
                    ),
                ),
                schema.Model.train_status == schema.Status.complete,
            )
        )

    if domain:
        results = results.filter(schema.Model.domain == domain)

    if username:
        results = results.filter(schema.Model.user.username == username)

    if access_level:
        conditions = []
        for access in access_level:
            conditions.append(schema.Model.access_level == access)

        # We have to unpack the conditions to be able to processed by `or_` function.
        results = results.filter(or_(*conditions))

    results = [get_high_level_model_info(result) for result in results]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the list",
        data=jsonable_encoder(results),
    )


@model_router.get("/name-check")
def check_model(
    name: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    Check if a model with the given name exists for the authenticated user.

    Parameters:
    - name: The name of the model to check.
    - session: The database session (dependency).
    - authenticated_user: The authenticated user (dependency).

    Returns:
    - A JSON response indicating if the model is present.
    """
    user: schema.User = authenticated_user.user
    model: schema.Model = (
        session.query(schema.Model)
        .filter(and_(schema.Model.name == name, schema.Model.user_id == user.id))
        .first()
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully checked for model name",
        data={"model_present": True if model else False},
    )


class SaveNDBDeployedModel(BaseModel):
    deployment_id: str
    model_id: str
    base_model_id: str
    model_name: str
    metadata: Dict[str, str]


@model_router.post("/save-deployed")
def save_deployed_model(
    body: SaveNDBDeployedModel,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    Save a deployed model.

    Parameters:
    - body: The body of the request containing deployment details.
    - session: The database session (dependency).
    - authenticated_user: The authenticated user (dependency).

    Returns:
    - A JSON response indicating the status of saving the model.
    """
    user: schema.User = authenticated_user.user
    base_model: schema.Model = session.query(schema.Model).get(body.base_model_id)
    user: schema.User = session.query(schema.User).get(user.id)

    new_model = schema.Model(
        id=body.model_id,
        name=body.model_name,
        train_status=schema.Status.complete,
        access_level=schema.Access.private,
        domain=user.email.split("@")[1],
        parent_deployment_id=body.deployment_id,
        parent_id=base_model.id,
        type=base_model.type,
        sub_type=base_model.sub_type,
        team_id=user.team_id,
    )

    session.add(new_model)
    session.commit()
    session.refresh(new_model)

    metadata: schema.MetaData = schema.MetaData(
        model_id=body.model_id, deployment=body.metadata
    )

    session.add(metadata)
    session.commit()

    return {"message": "successfully added the model."}


class ModelInfo(BaseModel):
    type: str
    sub_type: Optional[str] = None
    access_level: schema.Access = "public"
    metadata: Optional[Dict[str, str]] = None


@model_router.get("/pending-train-models")
def pending_train_models(
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    Get a list of all in progress or not started training models for the logged-in user.

    Returns:
    - JSONResponse: A list of models that are in progress or not started.
    """
    user: schema.User = authenticated_user.user

    pending_model_train: List[schema.Model] = (
        session.query(schema.Model)
        .filter(
            schema.Model.user_id == user.id,
            schema.Model.train_status.in_(
                [schema.Status.in_progress, schema.Status.not_started]
            ),
        )
        .all()
    )

    results = [
        {
            "model_name": result.name,
            "status": result.train_status,
            "username": user.username,
        }
        for result in pending_model_train
    ]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully fetched the pending list",
        data=jsonable_encoder(results),
    )


@model_router.get("/upload-token")
def upload_token(
    model_name: str,
    size: int,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    Generates a token for uploading a model to the platform.

    Parameters:
    - model_name: str - The name that the uploaded model will take in the platform.
        Example: "my_new_model"
    - size: int - The size of the model to be uploaded.
        Example: 150

    Returns:
    - JSONResponse: A token, which is used to upload chunks of a model.
    """
    user: schema.User = authenticated_user.user
    try:
        validate_name(model_name)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    model: schema.Model = get_model(
        session, username=user.username, model_name=model_name
    )

    if model:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"There is already a model saved with {model_name}",
        )

    model_id = str(uuid.uuid4())

    token = storage.create_upload_token(
        model_identifier=f"{user.username}/{model_name}",
        user_id=str(user.id),
        model_id=model_id,
        expiration_min=get_expiry_min(size),
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the upload token",
        data={"token": token},
    )


@model_router.post("/upload-chunk")
def upload_chunk(
    chunk: UploadFile,
    chunk_number: int,
    model_type: str = "ndb",
    compressed: bool = True,
    authorization: str = Header(None),
):
    """
    Uploads a chunk of a model.

    Parameters:
    - chunk: UploadFile - The raw bytes of the chunk.
        Example: UploadFile(file=BytesIO(b"chunk data"), filename="chunk1.zip")
    - chunk_number: int - The position of the chunk of the model that is being uploaded.
        Example: 1
    - model_type: str - The type of model being uploaded (default: "ndb").
        Example: "ndb"
    - compressed: bool - Whether the chunk is compressed (default: True).
        Example: True
    - authorization: str - Bearer token that contains the token generated from /upload-token.
        Example: "Bearer <token>"

    Returns:
    - JSONResponse: Success message if the chunk is uploaded successfully.
    """
    if authorization is None:
        return response(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header is missing",
            headers={"WWW-Authenticate": "Bearer"},
        )

    try:
        scheme, token = authorization.split()
        if scheme.lower() != "bearer":
            return response(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid Authentication Scheme",
                headers={"WWW-Authenticate": "Bearer"},
            )
        if not token:
            return response(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Authorization token is missing",
                headers={"WWW-Authenticate": "Bearer"},
            )
    except ValueError:
        return response(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authorization header format",
            headers={"WWW-Authenticate": "Bearer"},
        )

    try:
        payload = storage.verify_upload_token(token)
    except Exception as error:
        return response(status_code=status.HTTP_401_UNAUTHORIZED, message=str(error))

    try:
        chunk_data = chunk.file.read()
        storage.upload_chunk(
            payload["model_id"], chunk_data, chunk_number, model_type, compressed
        )
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="Uploaded chunk",
    )


@model_router.post("/upload-commit")
def upload_commit(
    total_chunks: int,
    body: ModelInfo,
    authorization: str = Header(None),
    session: Session = Depends(get_session),
):
    """
    Commits the upload of a model after all chunks have been uploaded.

    Parameters:
    - total_chunks: int - The total number of chunks uploaded.
        Example: 10
    - body: ModelInfo - The information about the model being uploaded.
        Example:
        ```
        {
            "type": "ndb",
            "sub_type": "subtype_example",
            "access_level": "public",
            "metadata": {
                "key1": "value1",
                "key2": "value2"
            }
        }
        ```
    - authorization: str - Bearer token that contains the token generated from /upload-token.
        Example: "Bearer <token>"

    Returns:
    - JSONResponse: Success message if the model upload is committed successfully.
    """
    if authorization is None:
        return response(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header is missing",
            headers={"WWW-Authenticate": "Bearer"},
        )

    try:
        scheme, token = authorization.split()
        if scheme.lower() != "bearer":
            return response(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid Authentication Scheme",
                headers={"WWW-Authenticate": "Bearer"},
            )
        if not token:
            return response(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Authorization token is missing",
                headers={"WWW-Authenticate": "Bearer"},
            )
    except ValueError:
        return response(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authorization header format",
            headers={"WWW-Authenticate": "Bearer"},
        )

    try:
        payload = storage.verify_upload_token(token)
    except Exception as error:
        return response(status_code=status.HTTP_401_UNAUTHORIZED, message=str(error))

    model_name = payload["model_identifier"].split("/")[1]
    model: schema.Model = (
        session.query(schema.Model)
        .filter(
            schema.Model.user_id == payload["user_id"], schema.Model.name == model_name
        )
        .first()
    )

    if model:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"There is already a model saved with {model_name}",
        )

    try:
        storage.commit_upload(payload["model_id"], total_chunks)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    user: schema.User = session.query(schema.User).get(payload["user_id"])
    domain = user.email.split("@")[1]

    try:
        new_model = schema.Model(
            id=payload["model_id"],
            name=model_name,
            access_level=body.access_level,
            type=body.type,
            sub_type=body.sub_type,
            domain=domain,
            user_id=payload["user_id"],
            train_status=schema.Status.complete,
        )

        session.add(new_model)
        session.commit()
        session.refresh(new_model)
    except Exception as err:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message=str(err),
        )

    if body.metadata:
        new_metadata = schema.MetaData(
            model_id=payload["model_id"],
            general=body.metadata,
        )

        session.add(new_metadata)
        session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Committed model",
    )


@model_router.get("/public-download")
def download_public_model(
    model_identifier: str,
    session: Session = Depends(get_session),
):
    """
    Downloads specified public model. No login required.

    Parameters:
    - model_identifier: str - The model identifier of the model to be downloaded.
        Example: "user123/my_model"

    Returns:
    - StreamingResponse: Streams the download of the model to the caller.
    """
    try:
        model: schema.Model = get_model_from_identifier(model_identifier, session)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    if model.access_level != "public":
        return response(
            status_code=status.HTTP_403_FORBIDDEN,
            message="You cannot access this model without login.",
        )

    model.downloads += 1
    session.commit()

    # Determine if the model should be compressed
    compressed = model.type == "ndb"
    storage.prepare_download(model.id, model_type=model.type, compressed=compressed)

    def iterfile():
        for chunk in storage.download_chunk_stream(
            model.id, block_size=8192, model_type=model.type, compressed=compressed
        ):
            yield chunk

    # Set the appropriate filename based on the model type and compression
    if compressed:
        filename = f"{model_identifier}.{model.type}.zip"
        media_type = "application/zip"
    else:
        filename = f"{model_identifier}.{model.type}"
        media_type = "application/octet-stream"

    headers = {"Content-Disposition": f'attachment; filename="{filename}"'}

    return StreamingResponse(iterfile(), headers=headers, media_type=media_type)


@model_router.get("/download")
def download_model(
    model_identifier: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    """
    Downloads specified model.

    Parameters:
    - model_identifier: str - The model identifier of the model to be downloaded.
        Example: "user123/my_model"

    Returns:
    - StreamingResponse: Streams the download of the model to the caller.
    """
    try:
        model: schema.Model = get_model_from_identifier(model_identifier, session)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    if not model_accessible(model, authenticated_user.user):
        return response(
            status_code=status.HTTP_403_FORBIDDEN,
            message="You don't have access to download this model.",
        )

    model.downloads += 1
    session.commit()

    # Determine if the model should be compressed
    compressed = model.type == "ndb"
    storage.prepare_download(model.id, model_type=model.type, compressed=compressed)

    def iterfile():
        for chunk in storage.download_chunk_stream(
            model.id, block_size=8192, model_type=model.type, compressed=compressed
        ):
            yield chunk

    # Set the appropriate filename based on the model type and compression
    if compressed:
        filename = f"{model_identifier}.{model.type}.zip"
        media_type = "application/zip"
    else:
        filename = f"{model_identifier}.{model.type}"
        media_type = "application/octet-stream"

    headers = {"Content-Disposition": f'attachment; filename="{filename}"'}

    return StreamingResponse(iterfile(), headers=headers, media_type=media_type)


@model_router.get("/team-models", dependencies=[Depends(team_admin_or_global_admin)])
def list_team_models(
    team_name: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user
    if user.role not in [schema.Role.team_admin, schema.Role.global_admin]:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You do not have the required permissions.",
        )

    team = session.query(schema.Team).filter(schema.Team.name == team_name).first()
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    results = (
        session.query(schema.Model)
        .filter(schema.Model.team_id == team.id)
        .options(joinedload(schema.Model.user))
        .all()
    )

    results = [get_high_level_model_info(result) for result in results]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the team models list",
        data=jsonable_encoder(results),
    )
