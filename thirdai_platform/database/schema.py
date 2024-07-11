import enum
import re
from datetime import datetime

from sqlalchemy import (
    JSON,
    BigInteger,
    Boolean,
    Column,
    DateTime,
    Float,
    ForeignKey,
    Integer,
    String,
    text,
)
from sqlalchemy.dialects.postgresql import ENUM, UUID
from sqlalchemy.orm import declarative_base, relationship, validates

SQLDeclarativeBase = declarative_base()


class Status(str, enum.Enum):
    not_started = "not_started"
    starting = "starting"
    in_progress = "in_progress"
    stopping = "stopping"
    complete = "complete"
    failed = "failed"


class Access(str, enum.Enum):
    public = "public"
    protected = "protected"
    private = "private"


class User(SQLDeclarativeBase):
    __tablename__ = "users"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    username = Column(String(50), nullable=False, unique=True)
    email = Column(String(254), nullable=False, unique=True)
    password_hash = Column(
        String, nullable=True
    )  # If NULL then its verified from some of the OAuth providers.
    verified = Column(Boolean, default=False)
    verification_token = Column(
        UUID(as_uuid=True),
        unique=True,
        server_default=text("gen_random_uuid()"),
    )

    models = relationship("Model", back_populates="user", cascade="all, delete-orphan")

    @validates("username")
    def validate_username(self, key, username):
        # allow only alphanumeric characters, underscores, and hyphens
        assert re.match(
            r"^[\w-]+$", username
        ), "Username should only contain alphanumeric characters, underscores, and hyphens"
        return username

    @property
    def domain(self) -> str:
        return self.email.split("@")[1]


class Model(SQLDeclarativeBase):
    __tablename__ = "models"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    name = Column(String, nullable=False)
    train_status = Column(ENUM(Status), nullable=False, default=Status.not_started)
    type = Column(String(256), nullable=False)
    # trained_on = Column(String, nullable=True)
    # time_taken = Column(BigInteger, nullable=True)
    # latency = Column(Float, nullable=True)
    # dataset_size = Column(BigInteger, nullable=True)
    # num_params = Column(BigInteger, nullable=True)
    downloads = Column(Integer, nullable=False, default=0)
    # size = Column(BigInteger, nullable=True)
    # size_in_memory = Column(BigInteger, nullable=True)
    # thirdai_version = Column(String, nullable=True)
    access_level = Column(ENUM(Access), nullable=False, default=Access.private)
    # description = Column(String, nullable=True)
    domain = Column(String, nullable=True)
    published_date = Column(
        DateTime, default=datetime.utcnow().isoformat(), nullable=True
    )

    user_id = Column(
        UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"), nullable=False
    )

    user = relationship("User", back_populates="models")

    meta_data = relationship(
        "MetaData", back_populates="model", cascade="all, delete-orphan"
    )

    @validates("name")
    def validate_model_name(self, key, name):
        # allow only alphanumeric characters, underscores, and hyphens
        assert re.match(
            r"^[\w-]+$", name
        ), "Model name should only contain alphanumeric characters, underscores, and hyphens"
        return name


class MetaData(SQLDeclarativeBase):
    __tablename__ = "metadata"

    public = Column(JSON, nullable=True)
    protected = Column(JSON, nullable=True)
    private = Column(JSON, nullable=True)

    model_id = Column(
        UUID(as_uuid=True),
        ForeignKey("models.id", ondelete="CASCADE"),
        primary_key=True,
    )

    model = relationship("Model", back_populates="meta_data")
