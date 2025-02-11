import ast
import io
import json
import logging
import os
import traceback
from pathlib import Path
from typing import Any, Dict, Optional

import fitz
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from platform_common.logging import setup_logger
from platform_common.ndb.ndbv2_parser import convert_to_ndb_doc
from platform_common.utils import response
from pydantic import BaseModel
from thirdai import neural_db_v2 as ndbv2

load_dotenv()


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
            status_code=400,
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
            status_code=400,
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


def highlight_v1(highlights, source: str) -> Optional[bytes]:
    highlights = ast.literal_eval(highlights)

    doc = fitz.open(source)
    for key, val in highlights.items():
        page = doc[key]
        blocks = page.get_text("blocks")
        for i, b in enumerate(blocks):
            if i in val:
                rect = fitz.Rect(b[:4])
                page.add_highlight_annot(rect)

    return doc.tobytes()


def highlight_v2(highlights, source: str) -> Optional[bytes]:
    highlights = ast.literal_eval(highlights)

    doc = fitz.open(source)
    for page, bounding_box in highlights:
        doc[page].add_highlight_annot(fitz.Rect(bounding_box))

    return doc.tobytes()


# NOTE: its definitely a leaky abstraction that the user has to pass in the full
# source path here. This endpoint is for the go deployment job since the go fitz
# library doesn't have the add_highlight_annot method. There are other ways to
# highlight pdfs in go but it was just easier to do things this way. The main
# difficulty is in supporting the v1 and v2 methods of providing highlight
# information. Ideally we standardize to the v2 highlighting with bounding boxes
# but that would take some time and Universe changes. For now, since its only the
# frontend calling this endpoint, its probably fine like this.
class HighlightRequest(BaseModel):
    source: str
    chunk_metadata: Dict[str, Any]


# TODO add permissions
@app.get("highlight-pdf")
def highlight_pdf(req: HighlightRequest):
    if "highlight" in req.metadata:
        pdf_bytes = highlight_v1(req.metadata["highlight"], req.source)
    elif "chunk_boxes" in req.metadata:
        pdf_bytes = highlight_v2(req.metadata["chunk_boxes"], req.source)
    else:
        raise HTTPException(status_code=400, detail="Invalid chunk metadata")

    buffer = io.BytesIO(pdf_bytes)
    headers = {"Content-Disposition": f'inline; filename="{Path(source).name}"'}
    return Response(buffer.getvalue(), headers=headers, media_type="application/pdf")


@app.get("/parse/health")
async def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn

    logger.info("Starting parsing service...")
    uvicorn.run(app, host="localhost", port=8000, log_level="info")
