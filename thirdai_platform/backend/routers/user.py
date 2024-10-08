from fastapi import APIRouter, Depends, HTTPException, status, Request
import pathlib
from fastapi.security import HTTPBasic, HTTPBasicCredentials
from sqlalchemy.orm import Session, joinedload
from pydantic import BaseModel
from database.session import get_session
from backend.utils import response
from auth.jwt import AuthenticatedUser, verify_access_token
from auth.identity_providers.factory import identity_provider
from database import schema
from typing import Optional, List
from backend.auth_dependencies import global_admin_only
from fastapi.encoders import jsonable_encoder
from fastapi.templating import Jinja2Templates
from auth.identity_providers.base import (
    AccountSignupBody,
    AdminRequest,
    VerifyResetPassword,
    AccessToken,
)

user_router = APIRouter()
basic_security = HTTPBasic()


root_folder = pathlib.Path(__file__).parent
template_directory = root_folder.joinpath("../templates/").resolve()
templates = Jinja2Templates(directory=template_directory)


@user_router.post("/email-signup-basic")
def email_signup(
    body: AccountSignupBody,
    session: Session = Depends(get_session),
):
    existing_user = (
        session.query(schema.User)
        .filter(
            (schema.User.email == body.email) | (schema.User.username == body.username)
        )
        .first()
    )

    if existing_user:
        if existing_user.email == body.email:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="There is already an account associated with this email.",
            )
        else:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="There is already a user associated with this username.",
            )

    return identity_provider.create_user(body, session)


@user_router.post("/email-login")
def email_login(
    credentials: HTTPBasicCredentials = Depends(basic_security),
    session: Session = Depends(get_session),
):
    """
    Handle email and password login using PostgreSQL.
    """
    try:
        # Authenticate using username and password
        user_id, access_token = identity_provider.authenticate_user(
            credentials.username, credentials.password, session
        )

        # Retrieve user from the local PostgreSQL database
        user = session.query(schema.User).filter(schema.User.id == user_id).first()
        if not user:
            return response(
                status_code=status.HTTP_404_NOT_FOUND,
                message="User not found in local database.",
            )

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully logged in using email and password.",
            data={
                "user": {
                    "username": user.username,
                    "email": user.email,
                    "user_id": str(user.id),
                },
                "access_token": access_token,
            },
        )
    except ValueError as e:
        return response(status_code=status.HTTP_401_UNAUTHORIZED, message=str(e))


@user_router.post("/email-login-with-keycloak")
def email_login_with_keycloak(
    access_token: AccessToken,
    session: Session = Depends(get_session),
):
    print(access_token.access_token)
    try:
        # Authenticate using the access token from Keycloak
        user_id, access_token = identity_provider.verify_idp_token(
            access_token.access_token, session
        )

        # Retrieve user from the local PostgreSQL database
        user = session.query(schema.User).filter(schema.User.id == user_id).first()
        if not user:
            return response(
                status_code=status.HTTP_404_NOT_FOUND,
                message="User not found in local database.",
            )

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully logged in using Keycloak token.",
            data={
                "user": {
                    "username": user.username,
                    "email": user.email,
                    "user_id": str(user.id),
                },
                "access_token": access_token,
            },
        )
    except ValueError as e:
        return response(status_code=status.HTTP_401_UNAUTHORIZED, message=str(e))


@user_router.post("/add-global-admin", dependencies=[Depends(global_admin_only)])
def add_global_admin(
    admin_request: AdminRequest,
    session: Session = Depends(get_session),
):
    user = (
        session.query(schema.User)
        .filter(schema.User.email == admin_request.email)
        .first()
    )

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found."
        )

    user.global_admin = True
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message=f"User {admin_request.email} has been promoted to global admin.",
    )


@user_router.post("/delete-global-admin", dependencies=[Depends(global_admin_only)])
def demote_global_admin(
    admin_request: AdminRequest,
    session: Session = Depends(get_session),
):
    user = (
        session.query(schema.User)
        .filter(schema.User.email == admin_request.email)
        .first()
    )

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found."
        )

    if not user.global_admin:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="User is not a global admin.",
        )

        # Check if there is more than one global admin
    another_admin_exists = (
        session.query(schema.User)
        .filter(schema.User.global_admin == True, schema.User.id != user.id)
        .first()
    )

    if not another_admin_exists:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="There must be at least one global admin.",
        )

    # Demote the user
    user.global_admin = False
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message=f"User {admin_request.email} has been demoted from global admin.",
    )


@user_router.delete("/delete-user")
def delete_user(
    admin_request: AdminRequest,
    session: Session = Depends(get_session),
):
    try:
        identity_provider.delete_user(admin_request.email, session)
        return response(
            status_code=status.HTTP_200_OK,
            message=f"User {admin_request.email} has been successfully deleted.",
        )
    except ValueError as e:
        return response(status_code=status.HTTP_404_NOT_FOUND, message=str(e))


@user_router.get("/all-users", dependencies=[Depends(global_admin_only)])
def list_all_users(session: Session = Depends(get_session)):
    users: List[schema.User] = (
        session.query(schema.User)
        .options(joinedload(schema.User.teams).joinedload(schema.UserTeam.team))
        .all()
    )

    users_info = [
        {
            "id": user.id,
            "username": user.username,
            "email": user.email,
            "global_admin": user.global_admin,
            "teams": [
                {
                    "team_id": user_team.team_id,
                    "team_name": user_team.team.name,
                    "role": user_team.role,
                }
                for user_team in user.teams
            ],
        }
        for user in users
    ]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the list of all users",
        data=jsonable_encoder(users_info),
    )


@user_router.get("/info")
def get_user_info(
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: Optional[schema.User] = (
        session.query(schema.User)
        .options(joinedload(schema.User.teams).joinedload(schema.UserTeam.team))
        .filter(schema.User.id == authenticated_user.user.id)
        .first()
    )

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )

    user_info = {
        "id": user.id,
        "username": user.username,
        "email": user.email,
        "global_admin": user.global_admin,
        "teams": [
            {
                "team_id": user_team.team_id,
                "team_name": user_team.team.name,
                "role": user_team.role,
            }
            for user_team in user.teams
        ],
    }

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully retrieved user information",
        data=jsonable_encoder(user_info),
    )


@user_router.get("/redirect-verify")
def redirect_email_verify(verification_token: str, request: Request):
    try:
        verify_url = identity_provider.redirect_verify(verification_token)
        context = {"request": request, "verify_url": verify_url}
        return templates.TemplateResponse("verify_email_sent.html", context=context)
    except Exception as e:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Failed to verify email: {str(e)}",
        )


@user_router.post("/email-verify")
def email_verify(verification_token: str, session: Session = Depends(get_session)):
    try:
        identity_provider.email_verify(verification_token)
        return response(
            status_code=status.HTTP_200_OK, message="Email verification successful."
        )
    except Exception as e:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=f"Verification failed: {str(e)}",
        )


@user_router.post("/new-password")
def reset_password(body: VerifyResetPassword, session: Session = Depends(get_session)):
    return identity_provider.reset_password(body, session)


@user_router.post("/get-all-idps")
def reset_password(session: Session = Depends(get_session)):
    identity_providers = identity_provider.get_all_idps()

    return response(
        status_code=status.HTTP_200_OK,
        message="Returning Identity Providers.",
        data=identity_providers,
    )
