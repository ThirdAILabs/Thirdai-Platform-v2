# auth/identity_providers/postgres_provider.py

from auth.identity_providers.base import (
    AbstractIdentityProvider,
    AccountSignupBody,
)
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError
from database import schema
import bcrypt
import uuid
from auth.jwt import create_access_token
from fastapi import HTTPException, status
from auth.jwt import verify_access_token, AuthenticatedUser
from sqlalchemy.orm import Session


class PostgresIdentityProvider(AbstractIdentityProvider):
    """
    PostgreSQL implementation of identity provider using the 'UserPostgresIdentityProvider' table.
    """

    def get_userinfo(self, token: str, session: Session):
        """
        Retrieve user information from the PostgreSQL database using the given token (access token).

        Args:
            token (str): The JWT used for authentication.
            session (Session): The database session.

        Returns:
            dict: A dictionary containing user information (e.g., user ID, email, username).
        """
        try:
            # Use the verify_access_token function to validate the token and get user info
            authenticated_user: AuthenticatedUser = verify_access_token(token, session)

            # Return the user info in a dictionary
            user = authenticated_user.user
            return {
                "id": user.id,
                "username": user.username,
                "email": user.email,
                "verified": True,
            }

        except HTTPException as e:
            # Handle any errors with token verification or user lookup
            raise HTTPException(
                status_code=(
                    e.status_code
                    if isinstance(e, HTTPException)
                    else status.HTTP_401_UNAUTHORIZED
                ),
                detail="Invalid or expired token",
            )

    def create_user(self, user_data: AccountSignupBody, session: Session):
        hashed_password = bcrypt.hashpw(
            user_data.password.encode("utf-8"), bcrypt.gensalt()
        ).decode("utf-8")

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

            return str(new_user.id)
        except IntegrityError:
            session.rollback()
            raise ValueError("User with this email or username already exists.")

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
        if user_identity:
            session.delete(user_identity)
            user = (
                session.query(schema.User)
                .filter(schema.User.id == user_identity.id)
                .first()
            )
            if user:
                session.delete(user)
            session.commit()
        else:
            raise ValueError("User not found in PostgreSQL.")

    def authenticate_user(
        self, username_or_email: str, password: str, session: Session
    ):
        user_identity = self.get_user(username_or_email, session)
        if not user_identity:
            raise ValueError("User not found.")

        if not user_identity.password_hash:
            raise ValueError("Password authentication not available for this user.")

        if not bcrypt.checkpw(
            password.encode("utf-8"), user_identity.password_hash.encode("utf-8")
        ):
            raise ValueError("Invalid password.")

        if not user_identity.verified:
            raise ValueError("User is not verified yet.")

        return str(user_identity.id), create_access_token(
            user_identity.id, expiration_min=120
        )
