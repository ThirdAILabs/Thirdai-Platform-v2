import os
from contextlib import contextmanager
import bcrypt

from database import schema
from database.schema import SQLDeclarativeBase as Base
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.exc import SQLAlchemyError


def hash_password(password: str):
    byte_password = password.encode("utf-8")
    salt = bcrypt.gensalt()
    return bcrypt.hashpw(byte_password, salt).decode()


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
engine = create_engine(db_uri, echo=echo)

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
    def add_admin(cls):
        with get_session() as session:
            domain = admin_mail.split("@")[1]

            organization = (
                session.query(schema.Organization)
                .filter(schema.Organization.domain == domain)
                .first()
            )
            if not organization:
                try:
                    organization = schema.Organization(
                        domain=domain, name=domain.split(".")[0]
                    )
                    session.add(organization)
                    session.commit()
                    session.refresh(organization)
                except SQLAlchemyError as e:
                    session.rollback()
                    raise ValueError(f"Error creating organization: {str(e)}")

            user: schema.User = (
                session.query(schema.User)
                .filter(schema.User.email == admin_mail)
                .first()
            )

            if not user:
                user = schema.User(
                    username=admin_username,
                    email=admin_mail,
                    password_hash=hash_password(admin_password),
                    verified=True,
                    role=schema.Role.admin,
                    organization_id=organization.id,
                )
                session.add(user)
                session.commit()
                session.refresh(user)
            else:
                user.role = schema.Role.admin
                user.organization_id = organization.id
                session.commit()


AdminAddition.add_admin()
