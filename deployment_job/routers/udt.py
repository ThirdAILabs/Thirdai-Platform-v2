from variables import GeneralVariables, TypeEnum
from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from permissions import Permissions
from utils import propagate_error, response
from pydantic_models.inputs import BaseQueryParams
from routers.model import get_model

udt_router = APIRouter()
permissions = Permissions()

general_variables = GeneralVariables.load_from_env()

@udt_router.post("/predict")
@propagate_error
def udt_query(
    base_params: BaseQueryParams,
    _=Depends(permissions.verify_read_permission),
):
    model = get_model()  
    params = base_params.dict()
    
    results = model.predict(**params)
    
    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=jsonable_encoder(results),
    )