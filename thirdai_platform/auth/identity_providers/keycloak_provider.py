from auth.identity_providers.base import (
    AbstractIdentityProvider,
    AccountSignupBody,
    VerifyResetPassword,
)
from auth.utils import keycloak_admin, keycloak_openid, get_token
from database import schema
from sqlalchemy.orm import Session
from fastapi import HTTPException, status
from typing import Optional
import uuid
from backend.utils import response
import os
from auth.identity_providers.utils import send_verification_mail


class KeycloakIdentityProvider(AbstractIdentityProvider):

    def create_user(self, user_data: AccountSignupBody, session):
        """
        Create a new user in Keycloak and send a verification email if required.

        - In the test environment, users are automatically verified.
        - For other environments, a verification email is sent using the Keycloak-generated token.
        """
        # Check if the user already exists in Keycloak
        existing_user_id = keycloak_admin.get_user_id(user_data.username)
        if existing_user_id:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="User with this email or username already exists.",
            )

        is_test_environment = os.getenv("TEST_ENVIRONMENT", "False") == "True"

        keycloak_user_id = keycloak_admin.create_user(
            {
                "username": user_data.username,
                "email": user_data.email,
                "enabled": True,
                "emailVerified": is_test_environment,
                "credentials": [
                    {
                        "type": "password",
                        "value": user_data.password,
                        "temporary": False,
                    }
                ],
            }
        )

        new_user = schema.User(
            id=keycloak_user_id,
            username=user_data.username,
            email=user_data.email,
        )
        session.add(new_user)
        session.commit()
        session.refresh(new_user)

        if not is_test_environment:
            self.trigger_keycloak_verification_email(
                keycloak_user_id, user_data.email, user_data.username
            )

        return str(new_user.id)

    def trigger_keycloak_verification_email(
        self, user_id: str, email: str, username: str
    ):
        """
        Request Keycloak to send an email verification and send the custom email via the custom mailer.
        """
        try:
            keycloak_admin.send_verify_email(user_id)

            base_url = "http://localhost:8180"
            verification_link = (
                f"{base_url}/realms/master/verify-email?user_id={user_id}"
            )

            send_verification_mail(email, verification_link, username)
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to send verification email: {str(e)}",
            )

    def email_verify(self, user_id: str):
        """
        This function is optional if we rely on Keycloak's internal verification. If necessary,
        it can be used to manually verify users from the local DB, but ideally, Keycloak handles it.
        """
        try:
            keycloak_admin.update_user(user_id, {"emailVerified": True})
            return response(
                status_code=status.HTTP_200_OK, message="Email verification successful."
            )
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to verify user email in Keycloak: {str(e)}",
            )

    def get_userinfo(self, token: str, session: Session):
        try:
            user_info = keycloak_openid.userinfo(token)
            keycloak_user_id = user_info.get("sub")

            if not keycloak_user_id:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid token: user ID not found in token",
                )

            # Fetch the user from the local DB
            user = (
                session.query(schema.User)
                .filter(schema.User.id == keycloak_user_id)
                .first()
            )

            if not user:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found in local DB",
                )

            return {
                "id": user.id,
                "username": user.username,
                "email": user.email,
                "verified": user.verified,
            }

        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Failed to retrieve user info: {str(e)}",
            )

    def reset_password(
        self,
        body: VerifyResetPassword,
        session: Session,
    ):
        user = (
            session.query(schema.User).filter(schema.User.email == body.email).first()
        )
        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="User not found in local DB",
            )

        keycloak_admin.set_user_password(
            user_id=user.id, password=body.new_password, temporary=False
        )

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully changed the password.",
        )

    def redirect_verify(self, verification_token: str, request):
        """
        Redirect to the Keycloak email verification endpoint.
        """
        base_url = "http://localhost:8180"
        verify_url = f"{base_url}/auth/realms/master/protocol/openid-connect/registrations?client_id=account&verification_token={verification_token}"

        return verify_url

    def get_user(self, username_or_email: str, session: Session):
        user = (
            session.query(schema.User)
            .filter(
                (schema.User.username == username_or_email)
                | (schema.User.email == username_or_email)
            )
            .first()
        )
        return user

    def delete_user(self, username_or_email: str, session: Session):
        user = self.get_user(username_or_email, session)
        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND, detail="User not found."
            )

        try:
            keycloak_admin.delete_user(user.id)
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to delete user in Keycloak: {str(e)}",
            )

        session.delete(user)
        session.commit()

    def authenticate_user(
        self, username_or_email: str, password: str, session: Session
    ):
        access_token = get_token("new-client", username_or_email, password)
        if not access_token:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid credentials."
            )

        try:
            user_info = keycloak_openid.userinfo(access_token)
            keycloak_user_id = user_info.get("sub")
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to retrieve user info: {str(e)}",
            )

        return keycloak_user_id, access_token

    def reset_password(
        self,
        body: VerifyResetPassword,
        session: Session,
    ):
        user = (
            session.query(schema.User).filter(schema.User.email == body.email).first()
        )
        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="User not found in local DB",
            )

        keycloak_admin.set_user_password(
            user_id=user.id, password=body.new_password, temporary=False
        )

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully changed the password.",
        )
