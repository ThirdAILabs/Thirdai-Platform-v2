import asyncio
import os
import time
import traceback
from functools import wraps
from typing import Optional

import thirdai
import uvicorn
from fastapi import APIRouter, Depends, FastAPI, status
from fastapi.encoders import jsonable_encoder
from fastapi.middleware.cors import CORSMiddleware
from models.ndb_models import ShardedNDB, SingleNDB
from permissions import Permissions
from pydantic_models.inputs import BaseQueryParams, NDBExtraParams
from utils import delete_job, response
from variables import GeneralVariables, TypeEnum

general_variables = GeneralVariables.load_from_env()

app = FastAPI(
    docs_url=f"/{general_variables.deployment_id}/docs",
    openapi_url=f"/{general_variables.deployment_id}/openapi.json",
)

if general_variables.license_key == "file_license":
    thirdai.licensing.set_path(
        os.path.join(general_variables.model_bazaar_dir, "license/license.serialized")
    )
else:
    thirdai.licensing.activate(general_variables.license_key)

router = APIRouter()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# The following logic will start a timer at this Fast API application's start up.
# after n minutes, this service will shut down, unless a function that is decorated
# with @reset_timer is called, in which case the timer restarts.
reset_event = asyncio.Event()


def reset_timer(endpoint_func):
    @wraps(endpoint_func)
    def wrapper(*args, **kwargs):
        response = endpoint_func(*args, **kwargs)
        reset_event.set()
        return response

    return wrapper


async def async_timer():
    while True:
        try:
            await asyncio.wait_for(
                reset_event.wait(), timeout=900
            )  # 15 minutes = 900 seconds
            reset_event.clear()  # clear the event if the endpoint was hit within the timeout period
        except asyncio.TimeoutError:
            # insert logic to cancel inference session
            response, job_id = delete_job(
                general_variables.deployment_id, general_variables.task_runner_token
            )
            if response.status_code == 200:
                print(f"Job {job_id} stopped successfully")
            else:
                print(
                    f"Failed to stop job {job_id}. Status code: {response.status_code}, Response: {response.text}"
                )

            reset_event.clear()


def propagate_error(func):
    @wraps(func)
    def method(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message=str(traceback.format_exc()),
                success=False,
            )

    return method


def get_model():
    if general_variables.type == TypeEnum.NDB:
        if general_variables.num_shards:
            return ShardedNDB()
        else:
            return SingleNDB()
    else:
        raise


permissions = Permissions()
model = get_model()


@router.post("/predict")
@propagate_error
def ndb_query(
    base_params: BaseQueryParams,
    ndb_params: Optional[NDBExtraParams] = NDBExtraParams(),
    _=Depends(permissions.verify_read_permission),
):
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


@router.on_event("startup")
async def startup_event():
    time.sleep(10)
    model.reporter.deploy_complete(general_variables.deployment_id)


app.include_router(router, prefix=f"/{general_variables.deployment_id}")


if __name__ == "__main__":
    uvicorn.run(app, host="localhost", port=8000)
