import os
import uuid
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
LLM_CACHE_JOB_ID = "llm-cache"
TELEMETRY_ID = "telemetry"


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
        model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
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


async def restart_llm_cache_job():
    nomad_endpoint = os.getenv("NOMAD_ENDPOINT")
    if nomad_job_exists(LLM_CACHE_JOB_ID, nomad_endpoint):
        delete_nomad_job(LLM_CACHE_JOB_ID, nomad_endpoint)
    cwd = Path(os.getcwd())
    platform = get_platform()
    return submit_nomad_job(
        nomad_endpoint=nomad_endpoint,
        filepath=str(cwd / "backend" / "nomad_jobs" / "llm_cache_job.hcl.j2"),
        platform=platform,
        port=None if platform == "docker" else get_empty_port(),
        tag=os.getenv("TAG"),
        registry=os.getenv("DOCKER_REGISTRY"),
        docker_username=os.getenv("DOCKER_USERNAME"),
        docker_password=os.getenv("DOCKER_PASSWORD"),
        image_name=os.getenv("LLM_CACHE_IMAGE_NAME"),
        model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
        share_dir=os.getenv("SHARE_DIR"),
        python_path=get_python_path(),
        llm_cache_app_dir=str(get_root_absolute_path() / "llm_cache_job"),
    )


async def restart_telemetry_jobs():
    """
    Restart the telemetry jobs.

    Returns:
    - Response: The response from the Nomad API.
    """
    nomad_endpoint = os.getenv("NOMAD_ENDPOINT")
    if nomad_job_exists(TELEMETRY_ID, nomad_endpoint):
        delete_nomad_job(TELEMETRY_ID, nomad_endpoint)

    cwd = Path(os.getcwd())
    response = submit_nomad_job(
        nomad_endpoint=nomad_endpoint,
        filepath=str(cwd / "backend" / "nomad_jobs" / "telemetry.hcl.j2"),
        VM_DATA_DIR=os.path.join(
            os.getenv("SHARE_DIR"), "monitoring-data", "victoriametric"
        ),
        LOKI_DATA_DIR=os.path.join(os.getenv("SHARE_DIR"), "monitoring-data", "loki"),
        dashboards=str(
            cwd.joinpath("..", "local_setup", "telemetry_dashboards").resolve()
        ),
        GRAFANA_DATA_DIR=os.path.join(
            os.getenv("SHARE_DIR"), "monitoring-data", "grafana"
        ),
    )
    if response.status_code != 200:
        raise Exception(f"{response.text}")
