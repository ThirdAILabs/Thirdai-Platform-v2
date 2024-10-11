from database import schema
from database.session import get_session
from contextlib import contextmanager
from auth.utils import keycloak_admin
from auth.jwt import identity_provider_type
from backend.utils import hash_password


class AdminAddition:
    @classmethod
    def add_admin(
        cls,
        admin_mail: str,
        admin_username: str,
        admin_password: str,
    ):
        """
        Add or update a global admin based on the current identity provider (Keycloak or Postgres).
        If Keycloak is used, Google as an identity provider can also be added or updated.
        """
        with contextmanager(get_session)() as session:
            user_id = None
            if identity_provider_type == "postgres":
                # Handle Postgres identity provider logic
                user_identity = (
                    session.query(schema.UserPostgresIdentityProvider)
                    .filter((schema.UserPostgresIdentityProvider.email == admin_mail))
                    .first()
                )

                if not user_identity:
                    hashed_password = hash_password(admin_password)

                    new_user_identity = schema.UserPostgresIdentityProvider(
                        username=admin_username,
                        email=admin_mail,
                        password_hash=hashed_password,
                        verified=True,
                    )
                    session.add(new_user_identity)
                    session.commit()
                    session.refresh(new_user_identity)

                    user_id = new_user_identity.id

            elif identity_provider_type == "keycloak":
                # Keycloak logic
                keycloak_user_id = keycloak_admin.get_user_id(admin_username)
                if keycloak_user_id:
                    keycloak_admin.update_user(
                        keycloak_user_id, {"email": admin_mail, "emailVerified": True}
                    )
                else:
                    keycloak_user_id = keycloak_admin.create_user(
                        {
                            "username": admin_username,
                            "email": admin_mail,
                            "enabled": True,
                            "emailVerified": True,
                            "credentials": [
                                {
                                    "type": "password",
                                    "value": admin_password,
                                    "temporary": False,
                                }
                            ],
                        }
                    )

                user_id = keycloak_user_id

            user = session.query(schema.User).filter(schema.User.id == user_id).first()

            # if not user:
            #     user = schema.User(
            #         id=keycloak_user_id,
            #         username=admin_username,
            #         email=admin_password,
            #     )

            #     user.global_admin = True
            #     session.add(user)
            #     session.commit()


AdminAddition.add_admin(
    admin_mail="kc_admin@mail.com",
    admin_username="kc_admin",
    admin_password="password",
)
