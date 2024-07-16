import datetime

from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse


def response(status_code: int, message: str, data={}, success: bool = None):
    if success is not None:
        status = "success" if success else "failed"
    else:
        status = "success" if status_code < 400 else "failed"
    return JSONResponse(
        status_code=status_code,
        content={"status": status, "message": message, "data": jsonable_encoder(data)},
    )


def now():
    return datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0)
