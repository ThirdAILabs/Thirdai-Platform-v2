from abc import ABC, abstractmethod
from sqlalchemy.orm import Session
from pydantic import BaseModel


class AccountSignupBody(BaseModel):
    username: str
    email: str
    password: str


class AdminRequest(BaseModel):
    email: str


class VerifyResetPassword(BaseModel):
    email: str
    reset_password_code: int
    new_password: str


class AbstractIdentityProvider(ABC):
    """
    Abstract base class for identity providers.
    """

    @abstractmethod
    def create_user(self, user_data: AccountSignupBody, session: Session):
        pass

    @abstractmethod
    def get_user(self, username_or_email: str, session: Session):
        pass

    @abstractmethod
    def delete_user(self, username_or_email: str, session: Session):
        pass

    @abstractmethod
    def authenticate_user(
        self, username_or_email: str, password: str, session: Session
    ):
        pass

    @abstractmethod
    def reset_password(
        self,
        body: VerifyResetPassword,
        session: Session,
    ):
        pass
