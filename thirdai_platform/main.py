from dotenv import load_dotenv

load_dotenv()

import fastapi
import uvicorn
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

if __name__ == "__main__":
    uvicorn.run(app, host="localhost", port=8000)
