from auth.identity_providers.keycloak_provider import KeycloakIdentityProvider
from auth.identity_providers.postgres_provider import PostgresIdentityProvider
import os
from database import schema
from auth.identity_providers.base import AccountSignupBody
from database.session import get_session
from contextlib import contextmanager
from auth.utils import keycloak_admin
from auth.jwt import identity_provider_type


def get_identity_provider(provider_type: str):
    if provider_type.lower() == "keycloak":
        return KeycloakIdentityProvider()
    elif provider_type.lower() == "postgres":
        return PostgresIdentityProvider()
    else:
        raise ValueError("Invalid provider type. Choose 'keycloak' or 'postgres'.")


identity_provider = get_identity_provider(identity_provider_type)


class AdminAddition:
    @classmethod
    def add_admin(cls, admin_mail: str, admin_username: str, admin_password: str):
        """
        Add or update a global admin based on the current identity provider (Keycloak or Postgres).
        This method assumes that the admin's mail, username, and password are the same.
        Role assignment is skipped for now.
        """
        with contextmanager(get_session)() as session:
            if identity_provider_type == "postgres":
                user = identity_provider.get_user(admin_username, session)
            elif identity_provider_type == "keycloak":
                # since we would already be initialzing keycloak with this user as admin user
                keycloak_user_id = keycloak_admin.get_user_id(admin_username)

                user = (
                    session.query(schema.User)
                    .filter(schema.User.id == keycloak_user_id)
                    .first()
                )

                if not user:
                    user = schema.User(
                        id=keycloak_user_id,
                        username=admin_username,
                        email=admin_password,
                    )

                    session.add(user)
                    session.commit()
                    session.refresh(user)

            if not user:
                # Create the user using the selected identity provider (Keycloak or Postgres)
                new_user_id = identity_provider.create_user(
                    AccountSignupBody(
                        username=admin_username,
                        email=admin_mail,
                        password=admin_password,
                    ),
                    session,
                )
                user = (
                    session.query(schema.User)
                    .filter(schema.User.id == new_user_id)
                    .first()
                )
                user.global_admin = True
                session.commit()
            else:
                user.global_admin = True
                session.commit()


AdminAddition.add_admin(
    admin_mail="kc_admin@mail.com", admin_username="kc_admin", admin_password="password"
)
