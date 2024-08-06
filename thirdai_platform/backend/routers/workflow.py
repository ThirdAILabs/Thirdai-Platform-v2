from typing import List

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import get_model_from_identifier, list_workflow_models, response
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from sqlalchemy.orm import Session

workflow_router = APIRouter()


@workflow_router.post("/create")
def create_workflow(
    name: str,
    type: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user

    workflow: schema.Workflow = (
        session.query(schema.Workflow).filter_by(name=name, user_id=user.id).first()
    )

    if workflow:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="Workflow with this name already exists for this user.",
        )

    new_workflow = schema.Workflow(
        name=name, type=type, user_id=user.id, status=schema.Status.in_progress
    )

    session.add(new_workflow)
    session.commit()
    session.refresh(new_workflow)

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully added the workflow",
        data={"workflow_id": str(new_workflow.id)},
    )


@workflow_router.post("/add-models")
def add_models(
    workflow_id: str,
    model_identifiers: List[str],
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user

    workflow: schema.Workflow = (
        session.query(schema.Workflow)
        .filter(schema.Workflow.id == workflow_id, schema.Workflow.user_id == user.id)
        .first()
    )

    if not workflow:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="Workflow not found.",
        )

    for model_identifier in model_identifiers:
        try:
            model: schema.Model = get_model_from_identifier(model_identifier, session)
        except Exception as error:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=str(error),
            )
        workflow_model = schema.WorkflowModel(
            workflow_id=workflow.id, model_id=model.id
        )
        session.add(workflow_model)

    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Models added to workflow successfully.",
        data={"models": jsonable_encoder(list_workflow_models(workflow=workflow))},
    )


@workflow_router.post("/delete-models")
def delete_models(
    workflow_id: str,
    model_identifiers: List[str],
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user

    workflow: schema.Workflow = (
        session.query(schema.Workflow)
        .filter(schema.Workflow.id == workflow_id, schema.Workflow.user_id == user.id)
        .first()
    )

    if not workflow:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="Workflow not found.",
        )

    for model_identifier in model_identifiers:
        try:
            model: schema.Model = get_model_from_identifier(model_identifier, session)
        except Exception as error:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=str(error),
            )

        workflow_model: schema.WorkflowModel = (
            session.query(schema.WorkflowModel)
            .filter(
                schema.WorkflowModel.workflow_id == workflow_id,
                schema.WorkflowModel.model_id == model.id,
            )
            .first()
        )

        if not workflow_model:
            return response(
                status_code=status.HTTP_404_NOT_FOUND,
                message=f"Model {model.id} not found in workflow.",
            )

        session.delete(workflow_model)

    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Models deleted from workflow successfully.",
        data={"models": jsonable_encoder(list_workflow_models(workflow=workflow))},
    )


@workflow_router.post("/validate")
def validate_workflow(
    workflow_id: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    pass


@workflow_router.post("/stop")
def validate_workflow(
    workflow_id: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    pass


@workflow_router.post("/start")
def validate_workflow(
    workflow_id: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    pass
