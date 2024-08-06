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
        f"Checking Global Admin: {current_user.is_global_admin()}, {current_user.email}"
    )
    if not current_user.is_global_admin():
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="The user doesn't have enough privileges",
        )
    return current_user


# Note: Following make sure the user is a global admin or a user is an admin to one team,
# However, this doesnot qurantees that the user is an admin to a particular team. That should
# be handled in the required function depending upon access depending upon the case.
def team_admin_or_global_admin(current_user: schema.User = Depends(get_current_user)):
    if current_user.is_global_admin() or current_user.is_team_admin_of_any_team():
        return current_user

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN,
        detail="The user doesn't have enough privileges",
    )


def verify_model_access(
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
    if permission is None:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="You do not have permission to access this model",
        )

    if permission == schema.Permission.read or permission == schema.Permission.write:
        return model

    raise HTTPException(
        status_code=status.HTTP_403_FORBIDDEN, detail="Access denied to the model"
    )
