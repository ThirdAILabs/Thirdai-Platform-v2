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
from auth.identity_providers.utils import send_verification_mail
from auth.identity_providers.utils import delete_all_models_for_user


class PostgresIdentityProvider(AbstractIdentityProvider):

    def create_user(self, user_data: AccountSignupBody, session: Session):
        hashed_password = hash_password(user_data.password)
        is_test_environment = os.getenv("TEST_ENVIRONMENT", "False") == "True"

        new_user_identity = schema.UserPostgresIdentityProvider(
            username=user_data.username,
            email=user_data.email,
            password_hash=hashed_password,
            verified=is_test_environment,
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

            if not is_test_environment:
                self.send_verification_mail(
                    new_user.email,
                    new_user_identity.verification_token,
                    new_user.username,
                )

            return response(
                status_code=status.HTTP_200_OK,
                message="Successfully signed up via email.",
                data={
                    "user": {
                        "username": new_user.username,
                        "email": new_user.email,
                        "user_id": str(new_user.id),
                    },
                },
            )

        except IntegrityError:
            session.rollback()
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="User with this email or username already exists.",
            )

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

    def redirect_verify(self, verification_token: str, request):
        """
        Redirect to email verification endpoint.
        """
        base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
        args = {"verification_token": verification_token}
        verify_url = urljoin(base_url, f"api/user/email-verify?{urlencode(args)}")

        return verify_url

    def get_user(self, username_or_email: str, session: Session):
        user = (
            session.query(schema.UserPostgresIdentityProvider)
            .filter(
                (schema.UserPostgresIdentityProvider.username == username_or_email)
                | (schema.UserPostgresIdentityProvider.email == username_or_email)
            )
            .first()
        )
        return user

    def delete_user(self, username_or_email: str, session: Session):
        user_identity = self.get_user(username_or_email, session)
        if not user_identity:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="User not found in PostgreSQL.",
            )

        delete_all_models_for_user(user, session)
        session.delete(user_identity)

        user = (
            session.query(schema.User)
            .filter(schema.User.id == user_identity.id)
            .first()
        )

        if user:
            session.delete(user)

        session.commit()

    def authenticate_user(
        self, username_or_email: str, password: str, session: Session
    ):
        user_identity = self.get_user(username_or_email, session)
        if not user_identity:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND, detail="User not found."
            )

        if not user_identity.password_hash:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Password authentication not available for this user.",
            )

        if not bcrypt.checkpw(
            password.encode("utf-8"), user_identity.password_hash.encode("utf-8")
        ):
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid password."
            )

        if not user_identity.verified:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="User is not verified yet.",
            )

        return str(user_identity.id), create_access_token(
            user_identity.id, expiration_min=120
        )

    def reset_password(
        self,
        body: VerifyResetPassword,
        session: Session,
    ):
        user_identity = self.get_user(body.email, session)

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

    def get_all_idps(self):
        return []
