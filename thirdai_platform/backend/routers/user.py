from backend.auth_dependencies import global_admin_only
from backend.mailer import Mailer
from backend.utils import hash_password, response
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, HTTPException, Request, status
from fastapi.encoders import jsonable_encoder
from fastapi.security import HTTPBasic, HTTPBasicCredentials
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel
from sqlalchemy.orm import Session, joinedload

from auth.utils import keycloak_openid, keycloak_admin, oauth2_scheme

user_router = APIRouter()


class AccountSignupBody(BaseModel):
    username: str
    email: str
    password: str


class AdminRequest(BaseModel):
    email: str


@user_router.post("/email-signup-basic")
def email_signup(
    body: AccountSignupBody,
    session: Session = Depends(get_session),
):
    """
    Sign up a new user with Keycloak and store additional info in the local DB.
    """
    keycloak_user_id = keycloak_admin.create_user(
        {
            "username": body.username,
            "email": body.email,
            "enabled": True,
            "emailVerified": True,  # TODO(pratik): remove it after testing
            "credentials": [
                {"type": "password", "value": body.password, "temporary": False}
            ],
        }
    )

    new_user = schema.User(
        id=keycloak_user_id,
        username=body.username,
        email=body.email,
    )

    session.add(new_user)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully signed up via email.",
        data={
            "user": {
                "username": new_user.username,
                "email": new_user.email,
                "user_id": str(new_user.id),
            }
        },
    )


@user_router.get("/email-login")
def login():
    # Keycloak would handle the login flow, hence no need for email-login
    return


@user_router.post("/add-global-admin", dependencies=[Depends(global_admin_only)])
def add_global_admin(
    admin_request: AdminRequest,
    session: Session = Depends(get_session),
):
    """
    Promote a user to global admin using Keycloak roles and update the local DB.
    """
    user = (
        session.query(schema.User)
        .filter(schema.User.email == admin_request.email)
        .first()
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found in local DB"
        )

    user.global_admin = True
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message=f"User {admin_request.email} promoted to global admin",
    )


@user_router.post("/delete-global-admin", dependencies=[Depends(global_admin_only)])
def demote_global_admin(
    admin_request: AdminRequest,
    session: Session = Depends(get_session),
):
    """
    Demote a global admin using Keycloak and update the local DB.
    """
    user = (
        session.query(schema.User)
        .filter(schema.User.email == admin_request.email)
        .first()
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found in local DB"
        )

    user.global_admin = False
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message=f"User {admin_request.email} demoted from global admin",
    )


@user_router.delete("/delete-user")
def delete_user(
    admin_request: AdminRequest,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(global_admin_only),
):
    """
    Delete a user from Keycloak and the local system.
    """

    user = (
        session.query(schema.User)
        .filter(schema.User.email == admin_request.email)
        .first()
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found in local DB"
        )

    keycloak_user_id = keycloak_admin.get_user_id(user.id)

    keycloak_admin.delete_user(keycloak_user_id)

    session.delete(user)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message=f"User {admin_request.email} successfully deleted",
    )


@user_router.get("/all-users", dependencies=[Depends(global_admin_only)])
def list_all_users(session: Session = Depends(get_session)):
    """
    List all users in the system along with their team memberships and roles.
    """
    users = (
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
    token: str = Depends(oauth2_scheme), session: Session = Depends(get_session)
):
    """
    Get detailed information about the authenticated user from Keycloak and the local DB.
    """
    print(token)
    user_info = keycloak_openid.userinfo(token)
    keycloak_user_id = user_info.get("sub")

    user = session.query(schema.User).filter(schema.User.id == keycloak_user_id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found in local DB"
        )

    user_info_formatted = {
        "id": user.id,
        "username": user_info.get("preferred_username"),
        "email": user_info.get("email"),
        "global_admin": user.is_global_admin,
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
        data=jsonable_encoder(user_info_formatted),
    )


class VerifyResetPassword(BaseModel):
    email: str
    new_password: str


@user_router.post("/new-password")
def reset_password(body: VerifyResetPassword):
    """
    Reset user password using Keycloak API.
    """
    user_id = keycloak_admin.get_user_id(body.email)
    keycloak_admin.set_user_password(
        user_id=user_id, password=body.new_password, temporary=False
    )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully changed the password.",
    )
