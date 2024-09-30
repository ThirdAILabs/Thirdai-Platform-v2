from auth.identity_providers.base import (
    AbstractIdentityProvider,
    AccountSignupBody,
    VerifyResetPassword,
)
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError
from database import schema
import bcrypt
import uuid
from auth.jwt import create_access_token, verify_access_token, AuthenticatedUser
from fastapi import HTTPException, status
from auth.jwt import AuthenticatedUser
from backend.utils import response, hash_password
from urllib.parse import urlencode, urljoin
import os
from backend.mailer import Mailer


class PostgresIdentityProvider(AbstractIdentityProvider):
    """
    PostgreSQL implementation of identity provider using the 'UserPostgresIdentityProvider' table.
    """

    def get_userinfo(self, token: str, session: Session):
        """
        Retrieve user information from the PostgreSQL database using the given token (access token).
        """
        try:
            # Validate the token and get user info
            authenticated_user: AuthenticatedUser = verify_access_token(token, session)
            user = authenticated_user.user
            return {
                "id": user.id,
                "username": user.username,
                "email": user.email,
                "verified": True,
            }

        except HTTPException as e:
            raise HTTPException(
                status_code=(
                    e.status_code
                    if isinstance(e, HTTPException)
                    else status.HTTP_401_UNAUTHORIZED
                ),
                detail="Invalid or expired token",
            )

    def create_user(self, user_data: AccountSignupBody, session: Session):
        hashed_password = hash_password(user_data.password)

        new_user_identity = schema.UserPostgresIdentityProvider(
            username=user_data.username,
            email=user_data.email,
            password_hash=hashed_password,
            verified=False,
            verification_token=str(uuid.uuid4()),
        )

        try:
            session.add(new_user_identity)
            session.commit()
            session.refresh(new_user_identity)

            new_user = schema.User(
                id=new_user_identity.id,
                username=user_data.username,
                email=user_data.email,
            )
            session.add(new_user)
            session.commit()
            session.refresh(new_user)

            # Send verification mail if not in test mode
            if os.getenv("TEST_ENVIRONMENT", "False") != "True":
                self.send_verification_mail(
                    new_user.email,
                    new_user_identity.verification_token,
                    new_user.username,
                )

            return str(new_user.id)

        except IntegrityError:
            session.rollback()
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="User with this email or username already exists.",
            )

    def send_verification_mail(
        self, email: str, verification_token: str, username: str
    ):
        """
        Send verification email to the user.
        """
        subject = "Verify Your Email Address"
        base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
        args = {"verification_token": verification_token}
        verify_link = urljoin(base_url, f"api/user/redirect-verify?{urlencode(args)}")
        body = f"<p>Please click the following link to verify your email address: <a href='{verify_link}'>verify</a></p>"

        Mailer(to=f"{username} <{email}>", subject=subject, body=body)

    def email_verify(self, verification_token: str, session: Session):
        """
        Verify the user's email with the provided token.
        """
        user_identity = (
            session.query(schema.UserPostgresIdentityProvider)
            .filter(
                schema.UserPostgresIdentityProvider.verification_token
                == verification_token
            )
            .first()
        )

        if not user_identity:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Invalid or expired token.",
            )

        user_identity.verified = True
        user_identity.verification_token = None
        session.commit()

        return response(
            status_code=status.HTTP_200_OK,
            message="Email verification successful.",
        )

    def redirect_verify(self, verification_token: str, request):
        """
        Redirect to email verification endpoint.
        """
        base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
        args = {"verification_token": verification_token}
        verify_url = urljoin(base_url, f"api/user/email-verify?{urlencode(args)}")

        return {"verify_url": verify_url}

    def reset_password(self, body: VerifyResetPassword, session: Session):
        """
        Reset password after verifying the reset code.
        """
        user_identity = (
            session.query(schema.UserPostgresIdentityProvider)
            .filter(schema.UserPostgresIdentityProvider.email == body.email)
            .first()
        )

        if not user_identity:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="This email is not registered with any account.",
            )

        if not user_identity.reset_password_code:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Click on forgot password to get verification code.",
            )

        if user_identity.reset_password_code != body.reset_password_code:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Entered wrong reset password code.",
            )

        user_identity.reset_password_code = None
        user_identity.password_hash = hash_password(body.new_password)
        session.commit()

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully changed the password.",
        )
