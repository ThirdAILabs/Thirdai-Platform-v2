import os


from database.schema import SQLDeclarativeBase as Base
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from database import schema
from contextlib import contextmanager
from auth.utils import keycloak_admin, identity_provider_type
from backend.utils import hash_password


db_uri = os.getenv("DATABASE_URI")
if db_uri is None:
    raise ValueError("No DATABASE_URI environment variable set")

admin_username = os.getenv("ADMIN_USERNAME")
admin_mail = os.getenv("ADMIN_MAIL")
admin_password = os.getenv("ADMIN_PASSWORD")

env_variable_names = [
    "ADMIN_USERNAME",
    "ADMIN_MAIL",
    "ADMIN_PASSWORD",
]

# Check if any of the environment variables are missing
missing_variables = [var for var in env_variable_names if os.getenv(var) is None]

if missing_variables:
    raise FileNotFoundError(
        f"The following environment variables are missing: {', '.join(missing_variables)}"
    )

# Determine the environment
app_env = os.getenv("APP_ENV", "Production")

# Set the echo parameter based on the environment
echo = True if app_env == "development" else False

# Create the SQLAlchemy engine
engine = create_engine(
    db_uri,
    echo=echo,  # Controls whether SQL statements are logged. Useful for debugging in development.
    pool_size=20,  # Number of connections to maintain in the pool for efficient connection reuse.
    max_overflow=30,  # Additional connections allowed beyond the pool size for handling spikes in load.
    pool_timeout=30,  # Maximum time (in seconds) to wait for a connection before timing out.
)

Session = sessionmaker(autocommit=False, autoflush=False, bind=engine)


"""Create all tables defined in the Base metadata."""
Base.metadata.create_all(engine, checkfirst=True)


def get_session():
    session = Session()
    try:
        yield session
    except Exception:
        session.rollback()
        raise
    finally:
        session.close()


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
                            "realmRoles": ["admin"],
                        }
                    )

            user = (
                session.query(schema.User)
                .filter(schema.User.email == admin_mail)
                .first()
            )

            if user:
                session.delete(user)
                session.commit()

            user = schema.User(
                id=keycloak_user_id,
                username=admin_username,
                email=admin_mail,
            )

            user.global_admin = True
            session.add(user)
            session.commit()


AdminAddition.add_admin(
    admin_mail=admin_mail,
    admin_username=admin_username,
    admin_password=admin_password,
)
