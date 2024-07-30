import hvac
from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import get_model_from_identifier, response
from database import schema
from database.session import get_session
from fastapi import Depends, HTTPException, status
from sqlalchemy.orm import Session
from database.session import get_session
from auth.jwt import AuthenticatedUser, verify_access_token
from database import schema
from thirdai_platform.backend.routers.utils import get_model_from_identifier, response
import hvac


def get_vault_client():
    # TODO(pratik): Change token in production environment
    # Note(pratik): Refer to the following for instruction to set up vault
    # https://waytohksharma.medium.com/install-hashicorp-vault-on-mac-fdbd8cd9113b
    client = hvac.Client(
        url="http://127.0.0.1:8200", token="00000000-0000-0000-0000-000000000000"
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


def global_admin_only(current_user: schema.User = Depends(get_current_user)):
    if current_user.role != schema.Role.global_admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="The user doesn't have enough privileges",
        )
    return current_user


def team_admin_or_global_admin(current_user: schema.User = Depends(get_current_user)):
    if current_user.role not in [schema.Role.team_admin, schema.Role.global_admin]:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="The user doesn't have enough privileges",
        )
    return current_user


def team_admin_only(current_user: schema.User = Depends(get_current_user)):
    if current_user.role != schema.Role.team_admin:
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
        if current_user.team_id != model_owner.team_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to the model",
            )

    elif model.user_id != current_user.id and (
        current_user.role != schema.Role.team_admin
        or current_user.team_id != model_owner.team_id
    ):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Access denied to the model"
        )

    return model
