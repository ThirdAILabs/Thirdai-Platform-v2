from fastapi import Depends, HTTPException, status
from sqlalchemy.orm import Session
from database.session import get_session
from auth.jwt import AuthenticatedUser, verify_access_token
from database import schema
from thirdai_platform.backend.routers.utils import get_model_from_identifier, response
import hvac


def get_vault_client():
    client = hvac.Client(
        url="http://127.0.0.1:8200", token="hvs.2Sia7aY91hwidc5vU0WpzVBI"
    )
    if not client.is_authenticated():
        raise HTTPException(status_code=500, detail="Vault authentication failed")
    return client


def get_current_user(
    session: Session = Depends(get_session),
    token: str = Depends(verify_access_token),
):
    user = session.query(schema.User).filter(schema.User.id == token.user.id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user"
        )
    return user


def verify_admin_access(current_user: schema.User = Depends(get_current_user)):
    if current_user.role != schema.Role.admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Admin privileges required"
        )
    return current_user


def verify_model_access(
    model_identifier: str,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    try:
        model: schema.Model = get_model_from_identifier(model_identifier, session)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )

    model_owner = (
        session.query(schema.User).filter(schema.User.id == model.user_id).first()
    )

    if not model_owner:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Model owner not found"
        )

    if model.access_level == schema.Access.public:
        if current_user.organization_id != model_owner.organization_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to the model",
            )

    elif model.user_id != current_user.id and (
        current_user.role != schema.Role.admin
        or current_user.organization_id != model_owner.organization_id
    ):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Access denied to the model"
        )
