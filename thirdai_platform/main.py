import sys

from dotenv import load_dotenv

load_dotenv()

import fastapi
import uvicorn

from backend.routers.deploy import deploy_router as deploy
from backend.routers.models import model_router as model
from backend.routers.train import train_router as train
from backend.routers.user import user_router as user
from backend.routers.data import data_router as data
from backend.utils import restart_generate_job
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
app.include_router(data, prefix="/api/data", tags=["data"])
app.include_router(train, prefix="/api/train", tags=["train"])
app.include_router(model, prefix="/api/model", tags=["model"])
app.include_router(deploy, prefix="/api/deploy", tags=["deploy"])


@app.on_event("startup")
async def startup_event():
    try:
        print("Starting Generation Job...")
        await restart_generate_job()
        print("Successfully started Generation Job!")
    except Exception as error:
        print(f"Failed to start the Generation Job : {error}", file=sys.stderr)


if __name__ == "__main__":
    uvicorn.run(app, host="localhost", port=8000)
