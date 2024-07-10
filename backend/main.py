import os
import pathlib
from fastapi import Depends, FastAPI, HTTPException
import fastapi
from fastapi.responses import FileResponse, JSONResponse
from starlette.staticfiles import StaticFiles
from pydantic import BaseModel
import requests
from fastapi.security import HTTPBasic, HTTPBasicCredentials
from requests.auth import HTTPBasicAuth

app = FastAPI()

root_folder = pathlib.Path(__file__).parent

static_path = root_folder.joinpath("../frontend/build").resolve()

# When you refresh a page on a single-page application (SPA) like React,
# the browser sends a request to the server for the refreshed URL.
# If your FastAPI server isn't configured to handle this route,
# it might result in a 404 error or unexpected behavior.


# In your FastAPI code, you need to configure it to serve the React
# frontend's index.html file for all routes. This is known as a catch-all route.
@app.exception_handler(404)
async def custom_404_handler(request, exception):
    if request.url.path.startswith("/api"):
        return JSONResponse(
            status_code=fastapi.status.HTTP_404_NOT_FOUND,
            content={"message": f"Requested path not found."},
        )
    path_to_index = os.path.join(static_path, "index.html")
    return FileResponse(path_to_index)

class JobRequest(BaseModel):
    name: str
    image: str
    command: str

@app.post("/submit_job")
def submit_job(job: JobRequest, credentials: HTTPBasicCredentials = Depends(HTTPBasic())):
    job_spec = {
        "Job": {
            "ID": job.name,
            "Name": job.name,
            "Type": "batch",
            "TaskGroups": [
                {
                    "Name": job.name,
                    "Tasks": [
                        {
                            "Name": job.name,
                            "Driver": "docker",
                            "Config": {
                                "image": job.image,
                                "args": [job.command]
                            }
                        }
                    ]
                }
            ]
        }
    }
    response = requests.post(
        "http://localhost:8080/api/v1/dags/temp/dagRuns",
        json={"conf": job_spec},
        auth=HTTPBasicAuth(credentials.username, credentials.password)
    )
    print("It went here")
    if response.status_code != 200:
        raise HTTPException(status_code=response.status_code, detail=response.text)
    return {"message": "Job submitted successfully", "dag_run_id": response.json().get("dag_run_id")}

app.mount("/", StaticFiles(directory=static_path, html=True), name="frontend")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
