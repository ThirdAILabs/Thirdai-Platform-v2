from dotenv import load_dotenv

load_dotenv()

import fastapi
import uvicorn
from backend.deploy import deploy_router as deploy
from backend.logger import logger_router as logger
from backend.models import model_router as model
from backend.train import train_router as train
from backend.user import user_router as user
from fastapi.middleware.cors import CORSMiddleware

app = fastapi.FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(user, prefix="/api/user", tags=["user"])
app.include_router(train, prefix="/api/train", tags=["train"])
app.include_router(model, prefix="/api/model", tags=["model"])
app.include_router(deploy, prefix="/api/deploy", tags=["deploy"])
app.include_router(logger, prefix="/api/logger", tags=["logger"])

if __name__ == "__main__":
    uvicorn.run(app, host="localhost", port=8000)
