import os
import pathlib
from urllib.parse import urlencode, urljoin

import bcrypt
from auth.jwt import create_access_token
from backend.mailer import Mailer
from backend.utils import hash_password, response, log_function_name
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, Request, status
from fastapi.security import HTTPBasic, HTTPBasicCredentials
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel
from sqlalchemy.orm import Session
from . import logger

user_router = APIRouter()
basic_security = HTTPBasic()

root_folder = pathlib.Path(__file__).parent
template_directory = root_folder.joinpath("templates/").resolve()
templates = Jinja2Templates(directory=template_directory)


class AccountSignupBody(BaseModel):
    username: str
    email: str
    password: str


def send_verification_mail(email: str, verification_token: str, username: str):
    subject = "Verify Your Email Address"
    # TODO(anyone) Get the base path using request.url rather than hard coding.
    base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
    args = {"verification_token": verification_token}
    verify_link = urljoin(base_url, f"api/user/redirect-verify?{urlencode(args)}")
    # TODO(anyone) Do we also need to send the link?
    body = "<p> Please click the following link to verify your email address: <a href={}> verify</a> </p>".format(
        verify_link
    )

    Mailer(
        to=f"{username} <{email}>",
        subject=subject,
        body=body,
    )

@log_function_name
@user_router.post("/email-signup-basic")
def email_signup(
    body: AccountSignupBody,
    session: Session = Depends(get_session),
):
    """
    The user sign-up endpoint is designed to ensure a seamless registration process by avoiding
    duplicate accounts and enforcing unique preferred names for each user. As an additional
    layer of security, it initiates a verification mail to authenticate user identities effectively.
    """
    user = session.query(schema.User).filter(schema.User.email == body.email).first()
    if user:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="There is already an account associated with this email.",
        )

    name = (
        session.query(schema.User).filter(schema.User.username == body.username).first()
    )

    if name:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message="There is already an user associated with this name.",
        )
    try:
        user = schema.User(
            username=body.username,
            email=body.email,
            password_hash=hash_password(body.password),
            verified=False,
        )

        session.add(user)
        session.commit()
        session.refresh(user)
        logger.info('Sent the verification mail')
        send_verification_mail(
            user.email,
            str(user.verification_token),
            user.username,
        )
    except Exception as err:
        return response(status_code=status.HTTP_400_BAD_REQUEST, message=str(err))

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully signed up via email.",
        data={
            "user": {
                "username": user.username,
                "email": user.email,
                "user_id": str(user.id),
            },
        },
    )

@log_function_name
@user_router.get("/redirect-verify")
def redirect_email_verify(verification_token: str, request: Request):
    """
    To verify user emails successfully, we require redirection to the "/email-verify" endpoint.
    However, direct redirection isn't feasible since browsers generally utilize HTTP GET requests
    for redirection. To overcome this, a temporary page is employed to redirect to "/email-verify"
    through an HTTP POST request, ensuring seamless and secure email verification for users.
    """
    base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
    args = {"verification_token": verification_token}
    verify_url = urljoin(base_url, f"api/user/email-verify?{urlencode(args)}")

    context = {"request": request, "verify_url": verify_url}
    logger.info('Redirected for email verification')
    return templates.TemplateResponse("verify_email_sent.html", context=context)

@log_function_name
@user_router.post("/email-verify")
def email_verify(verification_token: str, session: Session = Depends(get_session)):
    """
    Verifies the email with the verification token we send to the mail.
    """
    user: schema.User = (
        session.query(schema.User)
        .filter(schema.User.verification_token == verification_token)
        .first()
    )

    if not user:
        return {
            "message": "Token not found: this could be due to user already being verified or invalid token."
        }

    user.verified = True
    user.verification_token = None

    session.commit()
    logger.info('Email verification successful')
    return {"message": "Email verification successful."}

@log_function_name
@user_router.get("/email-login")
def email_login(
    credentials: HTTPBasicCredentials = Depends(basic_security),
    session: Session = Depends(get_session),
):
    """
    The user login process verifies the email-password association and requires email verification.
    Access is granted with an access token only to verified users, ensuring secure platform usage.
    """
    user: schema.User = (
        session.query(schema.User)
        .filter(schema.User.email == credentials.username)
        .first()
    )
    if not user:
        return response(
            status_code=status.HTTP_401_UNAUTHORIZED,
            message="User is not yet registered.",
        )
    byte_password = credentials.password.encode("utf-8")
    if not bcrypt.checkpw(byte_password, user.password_hash.encode("utf-8")):
        return response(
            status_code=status.HTTP_401_UNAUTHORIZED, message="Invalid password."
        )

    if not user.verified:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST, message="User is not verified yet."
        )
    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully logged in via email",
        data={
            "user": {
                "username": user.username,
                "email": user.email,
                "user_id": str(user.id),
            },
            "access_token": (create_access_token(user.id, expiration_min=120)),
            "verified": user.verified,
        },
    )
