import json
import os
import shutil
from typing import Dict, Optional

from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse


def model_bazaar_path():
    return (
        "/opt/model_bazaar" if os.path.exists("/.dockerenv") else os.getenv("SHARE_DIR")
    )


def response(
    status_code: int, message: str, data: Dict = {}, success: bool = None
) -> JSONResponse:
    """
    Creates a JSON response with a given status code, message, and data.

    Args:
        status_code (int): HTTP status code.
        message (str): Message to include in the response.
        data (Dict, optional): Data to include in the response. Defaults to {}.
        success (bool, optional): Indicates success or failure. Defaults to None.

    Returns:
        JSONResponse: The JSON response.
    """
    if success is not None:
        status = "success" if success else "failed"
    else:
        status = "success" if status_code < 400 else "failed"
    return JSONResponse(
        status_code=status_code,
        content={"status": status, "message": message, "data": jsonable_encoder(data)},
    )


def save_dict(write_to: str, **kwargs):
    with open(write_to, "w") as fp:
        json.dump(kwargs, fp, indent=4)


def load_dict(path: str):
    with open(path, "r") as fp:
        return json.load(fp)


def get_section(docs: str, header: str) -> str:
    sections = docs.split("---")
    for section in sections:
        if header in section:
            return section.strip()
    return "Documentation not found."


def disk_usage(path: Optional[str] = None):
    if path is None:
        path = model_bazaar_path()
    disk_stat = shutil.disk_usage(path)
    return {"total": disk_stat.total, "used": disk_stat.used, "free": disk_stat.free}
