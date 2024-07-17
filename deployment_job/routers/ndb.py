import traceback
import uuid
from typing import Optional

import thirdai
from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from permissions import Permissions
from pydantic_models import inputs
from pydantic_models.inputs import BaseQueryParams, NDBExtraParams
from routers.model import get_model
from utils import propagate_error, response, validate_name
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


@ndb_router.post("/associate")
@propagate_error
def ndb_associate(
    input: inputs.AssociateInput,
    token=Depends(permissions.verify_write_permission),
):
    model = get_model()
    model.associate(text_pairs=input.text_pairs, token=token)

    return response(status_code=status.HTTP_200_OK, message="Sucessfully associated")


@ndb_router.get("/sources")
@propagate_error
def get_sources(_=Depends(permissions.verify_read_permission)):
    model = get_model()
    sources = model.sources()
    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=sources,
    )


@ndb_router.post("/delete")
@propagate_error
def delete(input: inputs.DeleteInput, _=Depends(permissions.verify_write_permission)):
    model = get_model()
    model.delete(source_ids=input.source_ids)

    return response(
        status_code=status.HTTP_200_OK,
        message=f"{len(input.source_ids)} file(s) deleted",
        success=True,
    )


@ndb_router.post("/save")
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
        model.save_ndb(model_id=model_id)
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
