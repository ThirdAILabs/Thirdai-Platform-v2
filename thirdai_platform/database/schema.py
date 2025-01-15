import enum
import re
import secrets
from datetime import datetime, timedelta

from sqlalchemy import (
    ARRAY,
    JSON,
    Boolean,
    Column,
    DateTime,
    ForeignKey,
    Index,
    Integer,
    String,
    UniqueConstraint,
    text,
)
from sqlalchemy.dialects.postgresql import ENUM, UUID
from sqlalchemy.orm import declarative_base, relationship, validates

SQLDeclarativeBase = declarative_base()


class UDT_Task(str, enum.Enum):
    TEXT = "text"
    TOKEN = "token"


class Status(str, enum.Enum):
    not_started = "not_started"
    starting = "starting"
    in_progress = "in_progress"
    stopped = "stopped"
    complete = "complete"
    failed = "failed"


class Role(enum.Enum):
    user = "user"
    team_admin = "team_admin"
    global_admin = "global_admin"


class Access(enum.Enum):
    private = "private"
    protected = "protected"
    public = "public"

    def restrictiveness(self):
        order = {"public": 0, "protected": 1, "private": 2}
        return order[self.value]


class Permission(enum.Enum):
    read = "read"
    write = "write"

    def restrictiveness(self):
        order = {"read": 0, "write": 1}
        return order[self.value]


class Team(SQLDeclarativeBase):
    __tablename__ = "teams"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    name = Column(String(100), nullable=False, unique=True)

    users = relationship(
        "UserTeam", back_populates="team", cascade="all, delete-orphan"
    )
    models = relationship("Model", back_populates="team", cascade="all, delete-orphan")


class User(SQLDeclarativeBase):
    __tablename__ = "users"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    username = Column(String(50), nullable=False, unique=True)
    email = Column(String(254), nullable=False, unique=True)
    password_hash = Column(
        String, nullable=True
    )  # If NULL then it's verified from some of the OAuth providers.
    verified = Column(Boolean, default=False)
    verification_token = Column(
        UUID(as_uuid=True), unique=True, server_default=text("gen_random_uuid()")
    )

    # checks whether this user is global_admin or not
    global_admin = Column(Boolean, default=False, nullable=False)

    teams = relationship(
        "UserTeam", back_populates="user", cascade="all, delete-orphan"
    )
    models = relationship("Model", back_populates="user", cascade="all, delete-orphan")

    model_permissions = relationship(
        "ModelPermission", back_populates="user", cascade="all, delete-orphan"
    )

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

    def get_team_roles(self):
        return (
            {"team_id": user_team.team_id, "role": user_team.role}
            for user_team in self.teams
        )

    def is_global_admin(self):
        return self.global_admin

    def is_team_admin_of_any_team(self):
        return any(user_team.role == Role.team_admin for user_team in self.teams)

    def is_team_admin_of_team(self, team_id: UUID):
        return any(
            user_team.role == Role.team_admin and user_team.team_id == team_id
            for user_team in self.teams
        )


class PasswordReset(SQLDeclarativeBase):
    __tablename__ = "password_resets"

    user_id = Column(
        UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"), primary_key=True
    )
    reset_code_hash = Column(String, nullable=False)
    expiration_time = Column(DateTime, nullable=False)

    @staticmethod
    def generate_expiration_time(minutes: int = 15):
        return datetime.utcnow() + timedelta(minutes=minutes)

    @staticmethod
    def generate_reset_code(num: int = 6):
        # Use secrets to generate a secure 6-digit code
        return "".join([str(secrets.randbelow(10)) for _ in range(num)])

    def is_valid(self):
        return datetime.utcnow() < self.expiration_time


class UserTeam(SQLDeclarativeBase):
    __tablename__ = "user_teams"

    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), primary_key=True)
    team_id = Column(UUID(as_uuid=True), ForeignKey("teams.id"), primary_key=True)
    role = Column(ENUM(Role), nullable=False)

    user = relationship("User", back_populates="teams")
    team = relationship("Team", back_populates="users")


class Model(SQLDeclarativeBase):
    __tablename__ = "models"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    name = Column(String, nullable=False)
    train_status = Column(ENUM(Status), nullable=False, default=Status.not_started)
    deploy_status = Column(ENUM(Status), nullable=False, default=Status.not_started)
    type = Column(String(256), nullable=False)
    sub_type = Column(String(256), nullable=True)
    downloads = Column(Integer, nullable=False, default=0)
    access_level = Column(ENUM(Access), nullable=False, default=Access.private)
    domain = Column(String, nullable=True)
    published_date = Column(
        DateTime, default=datetime.utcnow().isoformat(), nullable=True
    )
    default_permission = Column(
        ENUM(Permission), nullable=False, default=Permission.read
    )

    parent_id = Column(
        UUID(as_uuid=True), ForeignKey("models.id", ondelete="SET NULL"), nullable=True
    )  # Not null if this model comes from starting training from a base model

    user_id = Column(
        UUID(as_uuid=True),
        ForeignKey("users.id", ondelete="CASCADE"),
        nullable=False,
    )
    team_id = Column(UUID(as_uuid=True), ForeignKey("teams.id"), nullable=True)

    user = relationship("User", back_populates="models")
    team = relationship("Team", back_populates="models")

    meta_data = relationship(
        "MetaData", back_populates="model", uselist=False, cascade="all, delete-orphan"
    )

    model_permissions = relationship(
        "ModelPermission", back_populates="model", cascade="all, delete-orphan"
    )

    attributes = relationship(
        "ModelAttribute", back_populates="model", cascade="all, delete-orphan"
    )

    dependencies = relationship(
        "ModelDependency",
        back_populates="model",
        cascade="all, delete-orphan",
        foreign_keys="ModelDependency.model_id",
    )

    used_by = relationship(
        "ModelDependency",
        back_populates="dependency",
        cascade="all, delete-orphan",
        foreign_keys="ModelDependency.dependency_id",
    )

    def get_train_job_name(self):
        return f"train-{self.id}-{self.type}-{self.sub_type}"

    def get_datagen_job_name(self):
        return f"GenerateData-{self.id}"

    def get_deployment_name(self):
        return f"deployment-{self.id}"

    def get_default_permission(self):
        return self.default_permission

    @validates("name")
    def validate_model_name(self, key, name):
        # allow only alphanumeric characters, underscores, and hyphens
        assert re.match(
            r"^[\w-]+$", name
        ), "Model name should only contain alphanumeric characters, underscores, and hyphens"
        return name

    @validates("access_level")
    def validate_access_level(self, key, access_level):
        # If access level is 'protected', ensure team_id is not None
        if access_level == Access.protected and self.team_id is None:
            raise ValueError("team_id cannot be None when access_level is 'protected'.")

        return access_level

    @validates("team_id")
    def validate_team_id(self, key, team_id):
        # For protected access, team_id should not be None
        if self.access_level == Access.protected and team_id is None:
            raise ValueError("team_id cannot be None when access_level is 'protected'.")

        return team_id

    def get_user_permission(self, user):
        # check whether we can find permission in explicit permissions first
        explicit_permission = next(
            (mp for mp in self.model_permissions if mp.user_id == user.id), None
        )
        if explicit_permission:
            return explicit_permission.permission

        if user.id == self.user_id or user.is_global_admin():
            return Permission.write

        if self.access_level == Access.protected:
            user_team = next(
                (ut for ut in user.teams if ut.team_id == self.team_id), None
            )
            if user_team:
                if user_team.role == Role.team_admin:
                    return Permission.write
                return self.get_default_permission()

        if self.access_level == Access.public:
            return self.get_default_permission()

        return None

    def get_owner_permission(self, user):
        if user.id == self.user_id or user.is_global_admin():
            return True

        if self.access_level == Access.protected:
            if user.is_team_admin_of_team(self.team_id):
                return True

        return False

    def get_attributes(self):
        return {attribute.key: attribute.value for attribute in self.attributes}

    __table_args__ = (
        Index("train_status_index", "train_status"),
        Index("model_identifier_index", "user_id", "name"),
        UniqueConstraint("user_id", "name"),
    )


class ModelAttribute(SQLDeclarativeBase):
    __tablename__ = "model_attributes"

    model_id = Column(
        UUID(as_uuid=True),
        ForeignKey("models.id", ondelete="CASCADE"),
        primary_key=True,
    )

    key = Column(String, primary_key=True)
    value = Column(String, nullable=True)

    model = relationship("Model", back_populates="attributes")

    __table_args__ = (Index("model_attribute", "model_id"),)


class ModelPermission(SQLDeclarativeBase):
    __tablename__ = "model_permissions"

    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), primary_key=True)
    model_id = Column(UUID(as_uuid=True), ForeignKey("models.id"), primary_key=True)
    permission = Column(ENUM(Permission), nullable=False)

    user = relationship("User", back_populates="model_permissions")
    model = relationship("Model", back_populates="model_permissions")


class MetaData(SQLDeclarativeBase):
    __tablename__ = "metadata"

    general = Column(JSON, nullable=True)
    train = Column(JSON, nullable=True)

    model_id = Column(
        UUID(as_uuid=True),
        ForeignKey("models.id", ondelete="CASCADE"),
        primary_key=True,
    )

    model = relationship("Model", back_populates="meta_data")


class ModelDependency(SQLDeclarativeBase):
    __tablename__ = "model_dependencies"

    model_id = Column(
        UUID(as_uuid=True),
        ForeignKey("models.id", ondelete="CASCADE"),
        primary_key=True,
    )
    dependency_id = Column(
        UUID(as_uuid=True),
        ForeignKey("models.id", ondelete="CASCADE"),
        primary_key=True,
    )

    model = relationship("Model", back_populates="dependencies", foreign_keys=model_id)
    dependency = relationship(
        "Model", back_populates="used_by", foreign_keys=dependency_id
    )

    __table_args__ = (
        Index("model_dependency_index", "model_id"),
        Index("dependency_model_index", "dependency_id"),
        UniqueConstraint("model_id", "dependency_id", name="unique_model_dependency"),
    )


class Catalog(SQLDeclarativeBase):
    __tablename__ = "catalog"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    name = Column(String(100), nullable=False)
    task = Column(ENUM(UDT_Task), nullable=False)
    num_generated_samples = Column(Integer)
    target_labels = Column(ARRAY(String), nullable=False)


class Level(str, enum.Enum):
    warning = "warning"
    error = "error"


class JobMessage(SQLDeclarativeBase):
    __tablename__ = "job_messages"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )

    model_id = Column(
        UUID(as_uuid=True), ForeignKey("models.id", ondelete="CASCADE"), index=True
    )
    timestamp = Column(
        DateTime, nullable=False, default=lambda: datetime.utcnow().isoformat()
    )
    job_type = Column(String(100), nullable=False)
    level = Column(ENUM(Level), nullable=False)
    message = Column(String)


class IntegrationType(str, enum.Enum):
    openai = "openai"
    self_hosted = "self_hosted"
    anthropic = "anthropic"
    cohere = "cohere"


# TODO(david) move api keys to vault?
class Integrations(SQLDeclarativeBase):
    __tablename__ = "integrations"

    id = Column(
        UUID(as_uuid=True), primary_key=True, server_default=text("gen_random_uuid()")
    )
    type = Column(ENUM(IntegrationType), nullable=False)
    data = Column(JSON, nullable=True)
