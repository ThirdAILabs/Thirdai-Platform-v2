from typing import Optional

from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from permissions import Permissions
from pydantic_models import inputs
from pydantic_models.inputs import BaseQueryParams, NDBExtraParams
from routers.model import get_model
from utils import propagate_error, response
from variables import GeneralVariables, TypeEnum

ndb_router = APIRouter()
permissions = Permissions()

general_variables = GeneralVariables.load_from_env()


@ndb_router.post("/predict")
@propagate_error
def ndb_query(
    base_params: BaseQueryParams,
    ndb_params: Optional[NDBExtraParams] = NDBExtraParams(),
    _=Depends(permissions.verify_read_permission),
):
    model = get_model()
    params = base_params.dict()
    if general_variables.type == TypeEnum.NDB:
        extra_params = ndb_params.dict(exclude_unset=True)
        params.update(extra_params)

    results = model.predict(**params)

    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=jsonable_encoder(results),
    )


@ndb_router.post("/upvote")
@propagate_error
def ndb_upvote(
    input: inputs.UpvoteInput, token=Depends(permissions.verify_write_permission)
):
    model = get_model()
    model.upvote(text_id_pairs=input.text_id_pairs, token=token)

    return response(status_code=status.HTTP_200_OK, message="Sucessfully upvoted")
