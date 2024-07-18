import traceback
import uuid

import thirdai
from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from permissions import Permissions
from pydantic_models import inputs
from pydantic_models.inputs import BaseQueryParams
from routers.model import get_model
from utils import Status, now, propagate_error, response, validate_files, validate_name
from variables import GeneralVariables, TypeEnum

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


@udt_router.post("/save")
def save(
    input: inputs.SaveModel,
    token=Depends(permissions.verify_read_permission),
    override_permission=Depends(permissions.get_owner_permission),
):
    model = get_model()
    model_id = general_variables.model_id
    if not input.override:
        model_id = str(uuid.uuid4())
        if not input.model_name:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Model name is required for new model.",
            )
        try:
            validate_name(input.model_name)
        except Exception:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Name must only contain alphanumeric characters, underscores (_), and hyphens (-). ",
            )
        is_model_present = model.reporter.check_model_present(token, input.model_name)
        if is_model_present:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Model name already exists, choose another one.",
            )
    else:
        if not override_permission:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="You dont have permissions to override this model.",
            )

    try:
        model.save_udt(model_id=model_id)
        if not input.override:
            model.reporter.save_model(
                access_token=token,
                deployment_id=general_variables.deployment_id,
                model_id=model_id,
                base_model_id=general_variables.model_id,
                model_name=input.model_name,
                metadata={"thirdai_version": str(thirdai.__version__)},
            )
    except Exception as err:
        traceback.print_exc()
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, message=str(err)
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully saved the model.",
        data={"new_model_id": model_id if not input.override else None},
    )
