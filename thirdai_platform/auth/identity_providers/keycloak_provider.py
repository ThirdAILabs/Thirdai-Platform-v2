from auth.identity_providers.base import (
    AbstractIdentityProvider,
    AccountSignupBody,
)
from auth.utils import keycloak_admin, keycloak_openid, get_token
from database import schema
from sqlalchemy.orm import Session
from fastapi import HTTPException, status
from typing import Optional
import uuid


class KeycloakIdentityProvider(AbstractIdentityProvider):

    def get_userinfo(self, token: str, session: Session):
        """
        Retrieve user information from Keycloak using the access token.

        Args:
            token (str): The JWT access token issued by Keycloak.
            session (Session): The database session for querying local user data.

        Returns:
            dict: A dictionary containing user information (e.g., user ID, email, username).
        """
        try:
            # Use Keycloak's OpenID Connect userinfo endpoint to retrieve user info
            user_info = keycloak_openid.userinfo(token)

            # Extract the user ID from the token ('sub' is the user ID field in Keycloak)
            keycloak_user_id = user_info.get("sub")
            if not keycloak_user_id:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid token: user ID not found in token",
                )

            # Query the local database to fetch more user info if needed
            user = (
                session.query(schema.User)
                .filter(schema.User.id == keycloak_user_id)
                .first()
            )

            if not user:
                # Optionally, you could create the user in the local DB if they don't exist
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="User not found in local DB",
                )

            # Return user info
            return {
                "id": user.id,
                "username": user.username,
                "email": user.email,
                "verified": True,
            }

        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Failed to retrieve user info: {str(e)}",
            )

    def create_user(self, user_data: AccountSignupBody, session: Session):
        existing_user_id = keycloak_admin.get_user_id(user_data.username)
        if existing_user_id:
            raise ValueError("User already exists with the same username.")

        # keep user verified for now
        keycloak_user_id = keycloak_admin.create_user(
            {
                "username": user_data.username,
                "email": user_data.email,
                "enabled": True,
                "emailVerified": True,
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

        return str(new_user.id)

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
            raise ValueError("User not found.")

        keycloak_admin.delete_user(user.id)

        session.delete(user)
        session.commit()

    def authenticate_user(
        self, username_or_email: str, password: str, session: Session
    ):
        access_token = get_token("new-client", username_or_email, password)
        if not access_token:
            raise ValueError("Invalid credentials.")

        user_info = keycloak_openid.userinfo(access_token)
        keycloak_user_id = user_info.get("sub")

        return keycloak_user_id, access_token
