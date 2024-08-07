import os

import hvac
from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import get_model_from_identifier, response
from database import schema
from database.session import get_session
from fastapi import Depends, HTTPException, status
from sqlalchemy.orm import Session


def get_vault_client():
    # TODO(pratik): Change token in production environment
    # Note(pratik): Refer to the following for instruction to set up vault
    # https://waytohksharma.medium.com/install-hashicorp-vault-on-mac-fdbd8cd9113b
    vault_url = os.getenv("HASHICORP_VAULT_ENDPOINT", "http://127.0.0.1:8200")
    vault_token = os.getenv(
        "HASHICORP_VAULT_TOKEN", "00000000-0000-0000-0000-000000000000"
    )

    client = hvac.Client(url=vault_url, token=vault_token)

    if not client.is_authenticated():
        raise HTTPException(status_code=500, detail="Vault authentication failed")

    return client


def get_current_user(
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    return authenticated_user.user


def global_admin_only(current_user: schema.User = Depends(get_current_user)):
    print(
        f"Checking Global Admin for {current_user.email}: {current_user.is_global_admin()}"
    )
    if not current_user.is_global_admin():
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="The user doesn't have enough privileges",
        )
    return current_user


def team_admin_or_global_admin(
    team_id: str,
    current_user: schema.User = Depends(get_current_user),
    session: Session = Depends(get_session),
):
    team = session.query(schema.Team).filter(schema.Team.id == team_id).first()
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    if current_user.is_global_admin() or current_user.is_team_admin_of_team(team.id):
        return current_user

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN,
        detail="The user doesn't have enough privileges",
    )


def is_model_owner(
    model_identifier: str,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    model = get_model_from_identifier(model_identifier, session)
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Model with identifier {model_identifier} not found",
        )

    if model.get_owner_permission(current_user):
        return True

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN,
        detail="You do not have owner permissions to this model",
    )


def verify_model_read_access(
    model_identifier: str,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    model = get_model_from_identifier(model_identifier, session)
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Model with identifier {model_identifier} not found",
        )

    permission = model.get_user_permission(current_user)
    if permission in [schema.Permission.read, schema.Permission.write]:
        return model

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN,
        detail="You do not have read access to this model",
    )


def verify_model_write_access(
    model_identifier: str,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    model = get_model_from_identifier(model_identifier, session)
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Model with identifier {model_identifier} not found",
        )

    permission = model.get_user_permission(current_user)
    if permission == schema.Permission.write:
        return model

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN,
        detail="You do not have write access to this model",
    )
