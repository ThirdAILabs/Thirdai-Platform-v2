import os
from pathlib import Path

from backend.utils import (
    delete_nomad_job,
    get_empty_port,
    get_platform,
    get_python_path,
    get_root_absolute_path,
    nomad_job_exists,
    submit_nomad_job,
)

GENERATE_JOB_ID = "llm-generation"
THIRDAI_PLATFORM_FRONTEND_ID = "thirdai-platform-frontend"


async def restart_generate_job():
    """
    Restart the LLM generation job.

    Returns:
    - Response: The response from the Nomad API.
    """
    nomad_endpoint = os.getenv("NOMAD_ENDPOINT")
    if nomad_job_exists(GENERATE_JOB_ID, nomad_endpoint):
        delete_nomad_job(GENERATE_JOB_ID, nomad_endpoint)
    cwd = Path(os.getcwd())
    platform = get_platform()
    return submit_nomad_job(
        nomad_endpoint=nomad_endpoint,
        filepath=str(cwd / "backend" / "nomad_jobs" / "generation_job.hcl.j2"),
        platform=platform,
        port=None if platform == "docker" else get_empty_port(),
        tag=os.getenv("TAG"),
        registry=os.getenv("DOCKER_REGISTRY"),
        docker_username=os.getenv("DOCKER_USERNAME"),
        docker_password=os.getenv("DOCKER_PASSWORD"),
        image_name=os.getenv("GENERATION_IMAGE_NAME"),
        python_path=get_python_path(),
        generate_app_dir=str(get_root_absolute_path() / "llm_generation_job"),
    )


async def restart_thirdai_platform_frontend():
    nomad_endpoint = os.getenv("NOMAD_ENDPOINT")
    if nomad_job_exists(THIRDAI_PLATFORM_FRONTEND_ID, nomad_endpoint):
        delete_nomad_job(THIRDAI_PLATFORM_FRONTEND_ID, nomad_endpoint)
    cwd = Path(os.getcwd())
    return submit_nomad_job(
        nomad_endpoint=nomad_endpoint,
        filepath=str(
            cwd / "backend" / "nomad_jobs" / "thirdai_platform_frontend.hcl.j2"
        ),
        public_model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
        openai_api_key=os.getenv("GENAI_KEY"),
        deployment_base_url=os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT"),
        thirdai_platform_base_url=os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT"),
        platform=get_platform(),
        tag=os.getenv("TAG"),
        registry=os.getenv("DOCKER_REGISTRY"),
        docker_username=os.getenv("DOCKER_USERNAME"),
        docker_password=os.getenv("DOCKER_PASSWORD"),
        image_name=os.getenv("FRONTEND_IMAGE_NAME"),
        # Model bazaar dockerfile does not include neuraldb_frontend code,
        # but app_dir is only used if platform == local.
        app_dir=str(get_root_absolute_path() / "frontend"),
    )