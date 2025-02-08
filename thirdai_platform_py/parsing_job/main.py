import json
import logging
import traceback
from pathlib import Path

from dotenv import load_dotenv
from platform_common.logging import setup_logger
from platform_common.ndb.ndbv2_parser import convert_to_ndb_doc
from platform_common.utils import response

load_dotenv()
import os
from typing import Any, Dict

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel
from thirdai import neural_db_v2 as ndbv2

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

model_bazaar_dir = os.getenv("MODEL_BAZAAR_DIR")
log_dir: Path = Path(model_bazaar_dir) / "logs"

setup_logger(log_dir=log_dir, log_prefix="parsing")
logger = logging.getLogger("parsing")


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


def write_doc_to_json(doc: ndbv2.Document, json_out_file: str):
    chunks = list(doc.chunks())
    data = {"document": chunks[0].document[0], "text": [], "metadata": []}
    for chunk in chunks:
        if chunk.metadata is not None:
            metadata = chunk.metadata.to_dict(orient="records")
        else:
            metadata = [{}] * len(chunk)
        data["metadata"].extend(metadata)
        data["text"].extend((chunk.text + " " + chunk.keywords).to_list())

    with open(json_out_file, "w") as f:
        json.dump(data, f)


class ParseRequest(BaseModel):
    upload_id: str
    filename: str
    options: Dict[str, Any] = {}


# TODO add permissions. What's the best way to do service-to-service communication?
# We could ping the permissions endpoint but we'd be calling this function once
# per doc. Better would be if this endpoint is not accessible via external calls
# and only accessible via internal calls
@app.post("/parse")
def parse_doc(req: ParseRequest):
    file_path = os.path.join(model_bazaar_dir, "uploads", req.upload_id, req.filename)
    if not os.path.exists(file_path):
        raise HTTPException(
            status_code=404,
            detail=f"File {file_path} does not exist",
        )

    try:
        doc = convert_to_ndb_doc(
            resource_path=file_path,
            display_path=file_path,
            doc_id=None,  # doc id is specified at insertion time
            metadata=None,  # metadata is specified at insertion time
            options=req.options,
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail="Error parsing document: " + str(e),
        )

    if doc is None:
        raise HTTPException(
            status_code=404,
            detail="Invalid extension for doc, please use one of .pdf, .csv, .docx, or .html",
        )

    parsed_doc_dir = os.path.join(model_bazaar_dir, "uploads", req.upload_id, "parsed")
    os.makedirs(parsed_doc_dir, exist_ok=True)
    input_filename_no_ext = Path(req.filename).stem
    json_out_file = os.path.join(parsed_doc_dir, f"{input_filename_no_ext}.json")

    try:
        write_doc_to_json(doc, json_out_file)
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail="Error writing parsed doc contents: " + str(e),
        )

    return response(status_code=200, message="Successfully parsed document")


@app.get("/parse/health")
async def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn

    logger.info("Starting parsing service...")
    uvicorn.run(app, host="localhost", port=8000, log_level="info")
