# routers/telemetry_router.py

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Any
import json
import os

class TelemetryEvent(BaseModel):
    UserAction: str
    UIComponent: str
    UI: str
    data: Any = None

class TelemetryEventPackage(BaseModel):
    UserName: str
    timestamp: str
    UserMachine: str
    event: TelemetryEvent

LOG_FILE_PATH = "telemetry_logs.json"

telemetry_router = APIRouter()

@telemetry_router.post("/record-event")
async def record_event(telemetry_package: TelemetryEventPackage):
    log_entry = telemetry_package.dict()

    # Ensure the log file exists
    if not os.path.isfile(LOG_FILE_PATH):
        with open(LOG_FILE_PATH, 'w') as f:
            json.dump([], f)

    # Append the log entry to the JSON file
    try:
        with open(LOG_FILE_PATH, 'r+') as f:
            logs = json.load(f)
            logs.append(log_entry)
            f.seek(0)
            json.dump(logs, f, indent=4)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to write log: {e}")

    return {"message": "Event recorded successfully"}
