from typing import Annotated, Optional, Union

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import get_high_level_model_info, response
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, Query, status
from fastapi.encoders import jsonable_encoder
from sqlalchemy import and_, or_
from sqlalchemy.orm import Session, joinedload

model_router = APIRouter()


@model_router.get("/public-list")
def list_public_models(
    name: str,
    domain: Optional[str] = None,
    username: Optional[str] = None,
    session: Session = Depends(get_session),
):
    """
    lists models which are public, for this endpoint we dont need login.
    """
    results = (
        session.query(schema.Model)
        .options(joinedload(schema.Model.user))
        .filter(
            schema.Model.name.ilike(f"%{name}%"),
            schema.Model.access_level == schema.Access.public,
        )
    )

    results = results.filter(schema.Model.train_status == schema.Status.complete)

    if domain:
        results = results.filter(schema.Model.domain == domain)

    if username:
        results = results.filter(schema.Model.user.username == username)

    results = [get_high_level_model_info(result) for result in results]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the public list",
        data=jsonable_encoder(results),
    )


@model_router.get("/list")
def list_models(
    name: str,
    domain: Optional[str] = None,
    username: Optional[str] = None,
    access_level: Annotated[Union[list[str], None], Query()] = None,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user: schema.User = authenticated_user.user

    """
    The list requests gets all the models available based on the given name by doing a fuzzy search.
    and models which are accessible for the user.
    """
    results = (
        session.query(schema.Model)
        .options(joinedload(schema.Model.user))
        .filter(
            schema.Model.name.ilike(f"%{name}%"),
            or_(
                # public
                schema.Model.access_level == schema.Access.public,
                # protected and matching domain
                and_(
                    schema.Model.access_level == schema.Access.protected,
                    schema.Model.domain == user.domain,
                ),
                # private and matching user or admin
                and_(
                    schema.Model.access_level == schema.Access.private,
                    or_(
                        schema.Model.user_id == user.id,
                        schema.User.id == user.id,
                    ),
                ),
            ),
            schema.Model.train_status == schema.Status.complete,
        )
    )

    if domain:
        results = results.filter(schema.Model.domain == domain)

    if username:
        results = results.filter(schema.Model.user.username == username)

    if access_level:
        conditions = []
        for access in access_level:
            conditions.append(schema.Model.access_level == access)

        # We have to unpack the conditions to be able to processed by `or_` function.
        results = results.filter(or_(*conditions))

    results = [get_high_level_model_info(result) for result in results]

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully got the list",
        data=jsonable_encoder(results),
    )
