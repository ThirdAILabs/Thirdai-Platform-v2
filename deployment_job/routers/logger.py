from fastapi import APIRouter, FastAPI
from pydantic import BaseModel
import uuid
import logging
import logging_loki

# Define a Pydantic model for the event
class Event(BaseModel):
    # define more custom keys here as required
    UserName: str
    timestamp: str
    UserMachine: str
    
    

logging_router = APIRouter()

# Loki configuration
logging_loki.emitter.LokiEmitter.level_tag = "level"

loki_handler = logging_loki.LokiHandler(
    url="http://localhost:80/loki/api/v1/push",
    version="1",
)
logger = logging.getLogger("action-logger")
logger.addHandler(loki_handler)
logger.setLevel(logging.DEBUG)

@logging_router.post("/log-event")
def log_event(
    session_id: uuid.UUID,
    event: Event
):
    logger.info(
        f"Session ID: {session_id}",
        extra=event.dict()  # Convert the event to a dictionary
    )
    return {"message": "Event logged successfully"}
