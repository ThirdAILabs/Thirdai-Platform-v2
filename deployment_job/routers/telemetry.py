import json
from typing import Any

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from routers.model import get_model


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
