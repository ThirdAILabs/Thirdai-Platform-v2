import datetime
import os
from typing import Union

import fastapi
import jwt
from auth.utils import CREDENTIALS_EXCEPTION, token_bearer
from database import schema
from database.session import get_session
from pydantic import BaseModel
from sqlalchemy.orm import Session


class TokenPayload(BaseModel):
    user_id: str
    exp: datetime.datetime


class AuthenticatedUser(BaseModel):
    user: schema.User
    exp: datetime.datetime

    class Config:
        arbitrary_types_allowed = True


def now_plus_minutes(minutes):
    return datetime.datetime.now(datetime.timezone.utc).replace(
        microsecond=0
    ) + datetime.timedelta(minutes=minutes)


def create_access_token(user_id, expiration_min=15):
    # This is a helper function which returns an access token with the given
    # expiration that has the given in email in the 'email' field of the payload.
    access_token_expires = now_plus_minutes(expiration_min)
    payload = TokenPayload(user_id=str(user_id), exp=access_token_expires)

    access_token = jwt.encode(
        payload=payload.dict(),
        key=os.getenv("JWT_SECRET"),
        algorithm="HS256",
    )
    return access_token


def verify_access_token(
    access_token: str = fastapi.Depends(token_bearer),
    session: Session = fastapi.Depends(get_session),
) -> AuthenticatedUser:
    # This function verifies that the given access token is valid and returns the
    # email from the payload if it is. Throws if the token is invalid. This function
    # should be used in the depends clause of any protected endpoint.
    #
    # The fastapi.Depends function is a type of dependency manager which will force
    # the input to meet the required args of the dependent function and then call
    # the dependent function before executing the main body of the function for
    # this endpoint. See the comment above for `token_bearer` to see what exactly
    # it does in this case.
    # Docs: https://fastapi.tiangolo.com/tutorial/dependencies/
    try:
        # This function automatically checks for token expiration:
        # https://pyjwt.readthedocs.io/en/stable/usage.html#expiration-time-claim-exp
        payload = TokenPayload(
            **jwt.decode(
                jwt=access_token,
                key=os.getenv("JWT_SECRET"),
                algorithms=["HS256"],
            )
        )
        if payload.user_id is None:
            raise CREDENTIALS_EXCEPTION
        user: schema.User = session.query(schema.User).get(payload.user_id)
        if not user:
            raise CREDENTIALS_EXCEPTION
        return AuthenticatedUser(user=user, exp=payload.exp)
    except jwt.PyJWTError:
        raise CREDENTIALS_EXCEPTION


def verify_access_token_no_throw(
    access_token: str = fastapi.Depends(token_bearer),
    session: Session = fastapi.Depends(get_session),
) -> Union[AuthenticatedUser, fastapi.HTTPException]:
    try:
        return verify_access_token(access_token, session)
    except fastapi.HTTPException as e:
        return e
