from typing import Dict, List, Optional

from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, status
from pydantic import BaseModel
from sqlalchemy.orm import Session

from .utils import response

logger_router = APIRouter()


class LogData(BaseModel):
    user_id: Optional[str]
    model_id: Optional[str]
    deployment_id: Optional[str]
    action: str
    train_samples: List[Dict[str, str]]
    used: bool


@logger_router.post("/log")
def log(
    log_data: LogData,
    session: Session = Depends(get_session),
):
    if log_data.deployment_id:
        deployment: schema.Deployment = (
            session.query(schema.Deployment)
            .filter(schema.Deployment.id == log_data.deployment_id)
            .first()
        )

        if not deployment:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="No deployment with this id",
            )

        deployment_id = deployment.id
        model_id = deployment.model_id

    elif log_data.model_id:
        model: schema.Model = (
            session.query(schema.Model)
            .filter(schema.Model.id == log_data.model_id)
            .first()
        )

        if not model:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST, message="No model with this id"
            )

        deployment_id = None
        model_id = model.id

    else:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="You must provide a deployment id or model id",
        )

    new_log = schema.Log(
        deployment_id=deployment_id,
        model_id=model_id,
        action=log_data.action,
        train_samples=log_data.train_samples,
        used=log_data.used,
        user_id=log_data.user_id,
    )

    session.add(new_log)
    session.commit()

    return response(status_code=status.HTTP_200_OK, message="Sucessfully logged.")
