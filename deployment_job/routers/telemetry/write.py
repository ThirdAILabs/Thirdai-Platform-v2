import json

from fastapi import APIRouter, HTTPException
from pydantic_models.inputs import TelemetryEventPackage
from routers.model import get_model

telemetry_write_router = APIRouter()


@telemetry_write_router.post("/record-event")
async def record_event(telemetry_package: TelemetryEventPackage):
    model = get_model()
    log_entry = telemetry_package.dict()

    # Append the log entry to the JSON file
    try:
        with open(model.telemetry_path, "r+") as f:
            logs = json.load(f)
            logs.append(log_entry)
            f.seek(0)
            json.dump(logs, f, indent=4)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to write log: {e}")

    return {"message": "Event recorded successfully"}
