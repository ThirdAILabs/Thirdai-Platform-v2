import datetime
import re
import traceback
from functools import wraps

import requests
from fastapi import status
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


def delete_job(deployment_id, task_runner_token):
    job_id = f"deployment-{deployment_id}"
    job_url = f"http://172.17.0.1:4646/v1/jobs/{job_id}"
    headers = {"X-Nomad-Token": task_runner_token}
    response = requests.delete(job_url, headers=headers)
    return response, job_id


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


def validate_name(name):
    regex_pattern = "^[\w-]+$"
    if not re.match(regex_pattern, name):
        raise ValueError("name is not valid")
