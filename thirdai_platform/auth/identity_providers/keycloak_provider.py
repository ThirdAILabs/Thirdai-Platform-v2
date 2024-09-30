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
from backend.utils import response, send_verification_mail
import os


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
            raise ValueError("User already exists with the same username.")

        is_test_environment = os.getenv("TEST_ENVIRONMENT", "False") == "True"

        # Create the user in Keycloak
        keycloak_user_id = keycloak_admin.create_user(
            {
                "username": user_data.username,
                "email": user_data.email,
                "enabled": True,
                "emailVerified": is_test_environment,  # Automatically verified in test environment
                "credentials": [
                    {
                        "type": "password",
                        "value": user_data.password,
                        "temporary": False,
                    }
                ],
            }
        )

        # Add the user to the local DB
        new_user = schema.User(
            id=keycloak_user_id,
            username=user_data.username,
            email=user_data.email,
        )
        session.add(new_user)
        session.commit()
        session.refresh(new_user)

        # If not in the test environment, request Keycloak to send a verification email
        if not is_test_environment:
            # Use Keycloak's endpoint to send a verification email
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
            # Trigger email verification via Keycloak
            keycloak_admin.send_verify_email(user_id)

            # Send the custom email with a verification link
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
