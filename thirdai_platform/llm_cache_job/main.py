import logging
import os
import traceback
from pathlib import Path

from fastapi import APIRouter, Depends, FastAPI, Request, status
from fastapi.encoders import jsonable_encoder
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from licensing.verify import verify_license
from llm_cache_job.cache import Cache, NDBSemanticCache
from llm_cache_job.permissions import Permissions
from platform_common.logging import setup_logger

app = FastAPI()
router = APIRouter()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


license_key = os.getenv("LICENSE_KEY")
verify_license.activate_thirdai_license(license_key)

model_bazaar_dir = os.getenv("MODEL_BAZAAR_DIR")
model_id = os.getenv("MODEL_ID")
model_dir = Path(model_bazaar_dir) / "models" / model_id
log_dir: Path = Path(model_bazaar_dir) / "logs" / model_id

setup_logger(log_dir=log_dir, log_prefix="llm-cache")

logger = logging.getLogger("llm-cache")

permissions = Permissions()

cache: Cache = NDBSemanticCache(model_dir=model_dir, logger=logger)


@app.middleware("http")
async def log_requests(request: Request, call_next):
    response = await call_next(request)

    logger.info(
        f"Request: {request.method}; URl: {request.url} - {response.status_code}"
    )

    return response


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    # Log the traceback
    error_trace = traceback.format_exc()
    logger.error(f"Exception occurred: {error_trace}")

    # Return the exact exception message in the response
    return JSONResponse(
        status_code=500,
        content={"detail": str(exc)},
    )


@router.get("/suggestions", dependencies=[Depends(permissions.verify_read_permission)])
def suggestions(query: str):
    result = cache.suggestions(query=query)

    logger.info(f"found {len(result)} suggestions for query={query}")

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"status": "success", "suggestions": jsonable_encoder(result)},
    )


@router.get("/query", dependencies=[Depends(permissions.verify_read_permission)])
def cache_query(query: str):
    result = cache.query(query=query)

    if result:
        logger.info(f"found cached result for query={query}")
    else:
        logger.info(f"found no cached result for query={query}")

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"status": "success", "cached_response": jsonable_encoder(result)},
    )


@router.post("/insert")
def cache_insert(
    query: str,
    llm_res: str,
    _: str = Depends(permissions.verify_temporary_cache_access_token),
):
    cache.insert(query=query, llm_res=llm_res)

    logger.info(f"cached query={query}")

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"status": "success", "message": "inserted response into cache"},
    )


@router.get("/token", dependencies=[Depends(permissions.verify_read_permission)])
def temporary_cache_token():
    logger.info(f"creating cache token")

    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={
            "status": "success",
            "access_token": permissions.create_temporary_cache_access_token(
                model_id=model_id
            ),
        },
    )


app.include_router(router, prefix="/cache")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="localhost", port=8000, log_level="info")
