import asyncio
import os
import time
from typing import Any

import uvicorn
from deployment_job.permissions import Permissions
from deployment_job.reporter import Reporter
from deployment_job.routers.enterprise_search import EnterpriseSearchRouter
from deployment_job.routers.ndb import NDBRouter
from deployment_job.routers.udt import UDTRouter
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from platform_common.pydantic_models.deployment import DeploymentConfig
from platform_common.pydantic_models.training import ModelType
from prometheus_client import make_asgi_app
from thirdai import licensing


def load_config():
    with open(os.getenv("CONFIG_PATH")) as file:
        return DeploymentConfig.model_validate_json(file.read())


config: DeploymentConfig = load_config()
reporter = Reporter(config.model_bazaar_endpoint)

licensing.activate(config.license_key)

Permissions.init(
    model_bazaar_endpoint=config.model_bazaar_endpoint, model_id=config.model_id
)

app = FastAPI(
    docs_url=f"/{config.model_id}/docs", openapi_url=f"/{config.model_id}/openapi.json"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


if config.model_options.model_type == ModelType.NDB:
    backend_router_factory = NDBRouter
elif config.model_options.model_type == ModelType.UDT:
    backend_router_factory = UDTRouter
elif config.model_options.model_type == ModelType.ENTERPRISE_SEARCH:
    backend_router_factory = EnterpriseSearchRouter
else:
    raise ValueError(f"Unsupported ModelType '{config.model_options.model_type}'.")


# We have a case where we copy the ndb model for base model training and
# read the model for deployment we face the open database issue.
max_retries = 2  # Total attempts including the initial one
retry_delay = 5  # Delay in seconds before retrying

for attempt in range(1, max_retries + 1):
    try:
        backend_router = backend_router_factory(config, reporter)
        break  # Exit the loop if model loading is successful
    except Exception as err:
        if attempt < max_retries:
            time.sleep(retry_delay)
        else:
            reporter.update_deploy_status(config.model_id, "failed")
            raise  # Optionally re-raise the exception if you want the application to stop


app.include_router(backend_router.router, prefix=f"/{config.model_id}")

app.mount("/metrics", make_asgi_app())


@app.exception_handler(404)
async def custom_404_handler(request: Request, exc: Any) -> JSONResponse:
    return JSONResponse(
        status_code=404,
        content={"message": f"Request path '{request.url.path}' doesn't exist"},
    )


@app.get("/")
async def homepage(request: Request) -> dict:
    return {"Deployment"}


@app.on_event("startup")
async def startup_event() -> None:
    """
    Event handler for application startup.
    """
    try:
        await asyncio.sleep(10)
        reporter.update_deploy_status(config.model_id, "complete")
    except Exception as e:
        reporter.update_deploy_status(config.model_id, "failed")
        print(f"Failed to startup the application, {e}")
        raise e  # Re-raise the exception to propagate it to the main block


if __name__ == "__main__":
    try:
        uvicorn.run(app, host="localhost", port=8000)
    except Exception as e:
        print(f"Uvicorn failed to start: {str(e)}")
        reporter.update_deploy_status(config.model_id, "failed")
