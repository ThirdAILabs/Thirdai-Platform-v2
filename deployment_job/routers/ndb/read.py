import io
from pathlib import Path
from typing import Optional

import fitz
from fastapi import APIRouter, Depends, Response, status
from fastapi.encoders import jsonable_encoder
from permissions import Permissions
from pydantic_models.inputs import BaseQueryParams, NDBExtraParams
from routers.model import get_model
from utils import (
    highlighted_pdf_bytes,
    new_pdf_chunks,
    old_pdf_chunks,
    propagate_error,
    response,
)
from variables import GeneralVariables

permissions = Permissions()
general_variables = GeneralVariables.load_from_env()

ndb_read_router = APIRouter()


@ndb_read_router.post("/predict")
@propagate_error
def ndb_query(
    base_params: BaseQueryParams,
    ndb_params: Optional[NDBExtraParams] = NDBExtraParams(),
    token: str = Depends(permissions.verify_read_permission),
):
    model = get_model()
    params = base_params.dict()
    extra_params = ndb_params.dict(exclude_unset=True)
    params.update(extra_params)

    params["token"] = token

    results = model.predict(**params)

    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=jsonable_encoder(results),
    )


@ndb_read_router.get("/sources")
@propagate_error
def get_sources(token: str = Depends(permissions.verify_read_permission)):
    model = get_model()
    sources = model.sources()

    return response(
        status_code=status.HTTP_200_OK,
        message="Successful",
        data=sources,
    )


@ndb_read_router.get("/highlighted-pdf")
@propagate_error
def highlighted_pdf(
    reference_id: int, token: str = Depends(permissions.verify_read_permission)
):
    model = get_model()
    reference = model.db._savable_state.documents.reference(reference_id)
    buffer = io.BytesIO(highlighted_pdf_bytes(reference))
    headers = {
        "Content-Disposition": f'inline; filename="{Path(reference.source).name}"'
    }
    return Response(buffer.getvalue(), headers=headers, media_type="application/pdf")


@ndb_read_router.get("/pdf-blob")
@propagate_error
def pdf_blob(source: str, token: str = Depends(permissions.verify_read_permission)):
    buffer = io.BytesIO(fitz.open(source).tobytes())
    headers = {"Content-Disposition": f'inline; filename="{Path(source).name}"'}
    return Response(buffer.getvalue(), headers=headers, media_type="application/pdf")


@ndb_read_router.get("/pdf-chunks")
@propagate_error
def pdf_chunks(
    reference_id: int, token: str = Depends(permissions.verify_read_permission)
):
    model = get_model()
    reference = model.db.reference(reference_id)
    chunks = new_pdf_chunks(model.db, reference)
    if not chunks:
        chunks = old_pdf_chunks(model.db, reference)
    if chunks:
        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=chunks,
        )
    return response(
        status_code=status.HTTP_400_BAD_REQUEST,
        message=f"Reference with id ${reference_id} is not a PDF.",
        data={},
    )
