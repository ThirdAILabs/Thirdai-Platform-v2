import json
import os
from pathlib import Path
from typing import List

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import (
    delete_nomad_job,
    get_empty_port,
    get_model_from_identifier,
    get_platform,
    get_python_path,
    get_root_absolute_path,
    list_workflow_models,
    response,
    submit_nomad_job,
)
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from sqlalchemy.orm import Session

from thirdai_platform.licensing.verify.verify_license import verify_license

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
        name=name, type=type, user_id=user.id, status=schema.Status.not_started
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
    workflow: schema.Workflow = session.query(schema.Workflow).get(workflow_id)

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
    workflow: schema.Workflow = session.query(schema.Workflow).get(workflow_id)

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
    workflow: schema.Workflow = session.query(schema.Workflow).get(workflow_id)

    if not workflow:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="Workflow not found.",
        )

    workflow_models: List[schema.WorkflowModel] = workflow.workflow_models

    if not workflow_models:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="No models found in the workflow.",
        )

    issues = []

    for workflow_model in workflow_models:
        model: schema.Model = workflow_model.model

        if model.train_status != schema.Status.complete:
            issues.append("Training is not complete.")

        if model.deploy_status != schema.Status.complete:
            issues.append("Deployment is not complete.")

    if issues:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="Validation failed. Some models have issues.",
            data={"models": list_workflow_models(workflow)},
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="All models are properly trained and deployed.",
        data={"models": list_workflow_models(workflow)},
    )


@workflow_router.post("/stop")
def stop_workflow(
    workflow_id: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    workflow: schema.Workflow = session.query(schema.Workflow).get(workflow_id)

    if not workflow:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="Workflow not found.",
        )

    workflow_models: List[schema.WorkflowModel] = workflow.workflow_models

    for workflow_model in workflow_models:
        model: schema.Model = workflow_model.model
        if model.deploy_status in [schema.Status.complete, schema.Status.in_progress]:
            active_workflows_using_model = (
                session.query(schema.WorkflowModel)
                .join(schema.Workflow)
                .filter(
                    schema.WorkflowModel.model_id == model.id,
                    schema.Workflow.id != workflow_id,
                    schema.Workflow.status.in_(
                        [schema.Status.in_progress, schema.Status.complete]
                    ),
                )
                .count()
            )

            if active_workflows_using_model == 0:
                try:
                    delete_nomad_job(
                        job_id=f"deployment-{str(model.id)}",
                        nomad_endpoint=os.getenv("NOMAD_ENDPOINT"),
                    )
                    model.deploy_status = schema.Status.stopped
                    session.commit()
                except Exception as err:
                    workflow.status = schema.Status.stopped
                    session.commit()
                    return response(
                        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                        message=f"Failed to undeploy model {model.name}: {str(err)}",
                    )

    workflow.status = schema.Status.stopped
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Workflow stopped successfully.",
    )


@workflow_router.post("/start")
def start_workflow(
    workflow_id: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    workflow: schema.Workflow = session.query(schema.Workflow).get(workflow_id)

    if not workflow:
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="Workflow not found.",
        )

    workflow.status = schema.Status.in_progress
    session.commit()

    workflow_models: List[schema.WorkflowModel] = workflow.workflow_models

    if not workflow_models:
        workflow.status = schema.Status.stopped
        session.commit()
        return response(
            status_code=status.HTTP_404_NOT_FOUND,
            message="No models found in the workflow.",
        )

    all_training_complete = True

    for workflow_model in workflow_models:
        model: schema.Model = workflow_model.model

        if model.train_status != schema.Status.complete:
            all_training_complete = False

    if not all_training_complete:
        workflow.status = schema.Status.stopped
        session.commit()
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="Cannot start workflow. Some models are not ready.",
            data={"models": list_workflow_models(workflow)},
        )

    for workflow_model in workflow_models:
        model: schema.Model = workflow_model.model
        if model.deploy_status not in [
            schema.Status.starting,
            schema.Status.in_progress,
            schema.Status.complete,
        ]:
            try:
                meta_data = json.loads(model.meta_data.train)
                size_in_memory = int(meta_data["size_in_memory"])
            except (json.JSONDecodeError, KeyError) as e:
                workflow.status = schema.Status.stopped
                session.commit()
                raise Exception(
                    "Failed to parse model metadata or missing 'size_in_memory'."
                )
            memory = (size_in_memory // 1000000) + 1000  # MB required for deployment

            try:
                work_dir = os.getcwd()
                platform = get_platform()

                submit_nomad_job(
                    str(
                        Path(work_dir)
                        / "backend"
                        / "nomad_jobs"
                        / "deployment_job.hcl.j2"
                    ),
                    nomad_endpoint=os.getenv("NOMAD_ENDPOINT"),
                    platform=platform,
                    tag=os.getenv("TAG"),
                    registry=os.getenv("DOCKER_REGISTRY"),
                    docker_username=os.getenv("DOCKER_USERNAME"),
                    docker_password=os.getenv("DOCKER_PASSWORD"),
                    image_name=os.getenv("DEPLOY_IMAGE_NAME"),
                    port=None if platform == "docker" else get_empty_port(),
                    deployment_app_dir=str(get_root_absolute_path() / "deployment_job"),
                    model_id=str(model.id),
                    model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
                    share_dir=os.getenv("SHARE_DIR", None),
                    license_key=verify_license(
                        os.getenv(
                            "LICENSE_PATH",
                            "/model_bazaar/license/ndb_enterprise_license.json",
                        )
                    )["boltLicenseKey"],
                    genai_key=(os.getenv("GENAI_KEY", "")),
                    autoscaling_enabled="false",
                    autoscaler_max_count="1",
                    memory=memory,
                    type=model.type,
                    sub_type=model.sub_type,
                    python_path=get_python_path(),
                    aws_access_key=(os.getenv("AWS_ACCESS_KEY", "")),
                    aws_access_secret=(os.getenv("AWS_ACCESS_SECRET", "")),
                )

            except Exception as err:
                model.deploy_status = schema.Status.failed
                workflow.status = schema.Status.stopped
                session.commit()
                raise Exception(str(err))

    return response(
        status_code=status.HTTP_202_ACCEPTED,
        message="Workflow started successfully.",
        data={"models": list_workflow_models(workflow)},
    )
