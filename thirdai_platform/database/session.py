import os
from contextlib import contextmanager

from backend.utils import hash_password
from database import schema
from database.schema import SQLDeclarativeBase as Base
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

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


from auth.utils import keycloak_admin


class AdminAddition:
    @classmethod
    def add_admin(cls, admin_mail: str, admin_username: str, admin_password: str):
        # Check if the user exists in Keycloak by username
        existing_user = keycloak_admin.get_user_id(admin_username)

        if not existing_user:
            # Create the user in Keycloak if they don't exist
            new_user_id = keycloak_admin.create_user(
                {
                    "email": admin_mail,
                    "username": admin_username,
                    "enabled": True,
                    "credentials": [
                        {
                            "value": admin_password,
                            "type": "password",
                            "temporary": False,
                        }
                    ],
                    "emailVerified": True,
                }
            )

            if not new_user_id:
                raise Exception("Failed to create user in Keycloak")

            keycloak_user_id = keycloak_admin.get_user_id(admin_username)

            # Assign the 'global_admin' role to the new user
            global_admin_role = keycloak_admin.get_realm_role("global_admin")
            keycloak_admin.assign_realm_roles(
                user_id=keycloak_user_id, roles=[global_admin_role]
            )
        else:
            keycloak_user_id = existing_user
            # Assign 'global_admin' role if the user already exists in Keycloak
            global_admin_role = keycloak_admin.get_realm_role("global_admin")
            keycloak_admin.assign_realm_roles(
                user_id=keycloak_user_id, roles=[global_admin_role]
            )

        # Add or update the user in your application's database
        with contextmanager(get_session)() as session:
            user: schema.User = (
                session.query(schema.User)
                .filter(schema.User.email == admin_mail)
                .first()
            )

            if not user:
                # If the user does not exist in the database, add them with the global admin flag
                user = schema.User(
                    username=admin_username,
                    email=admin_mail,
                    id=keycloak_user_id,
                    global_admin=True,
                )
                session.add(user)
                session.commit()
                session.refresh(user)
            else:
                # If the user already exists, just update their role to global admin
                user.global_admin = True
                session.commit()


AdminAddition.add_admin("admin@example.com", "admin", "admin_password")
