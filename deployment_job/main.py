import asyncio
import time
from functools import wraps
from multiprocessing import Process

import uvicorn
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from reporter import Reporter
from routers.ndb import ndb_router, process_tasks
from routers.udt import udt_router
from utils import delete_job
from variables import GeneralVariables, TypeEnum

general_variables = GeneralVariables.load_from_env()

app = FastAPI(
    docs_url=f"/{general_variables.deployment_id}/docs",
    openapi_url=f"/{general_variables.deployment_id}/openapi.json",
)

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


if general_variables.type == TypeEnum.NDB:
    app.include_router(ndb_router, prefix=f"/{general_variables.deployment_id}")
elif general_variables.type == TypeEnum.UDT:
    app.include_router(udt_router, prefix=f"/{general_variables.deployment_id}")


@app.exception_handler(404)
async def custom_404_handler(request: Request, exc):
    return JSONResponse(
        status_code=404,
        content={"message": f"Request path '{request.url.path}' doesn't exist"},
    )


@app.get("/")
async def homepage(request: Request):
    return {"Deployment"}


@app.on_event("startup")
async def startup_event():
    time.sleep(10)
    reporter = Reporter(general_variables.model_bazaar_endpoint)
    reporter.deploy_complete(general_variables.deployment_id)

    if general_variables.type == TypeEnum.NDB:
        Process(target=process_tasks, daemon=True).start()


if __name__ == "__main__":
    uvicorn.run(app, host="localhost", port=8000)
