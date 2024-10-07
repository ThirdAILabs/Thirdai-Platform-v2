import json
import os
import traceback
from pathlib import Path
from typing import Optional, Type

from auth.jwt import verify_access_token
from backend.utils import (
    delete_nomad_job,
    get_platform,
    get_python_path,
    get_root_absolute_path,
    model_bazaar_path,
    nomad_job_exists,
    response,
    submit_nomad_job,
)
from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel, Field

recovery_router = APIRouter()


# Base configuration model for backup
class BackupConfig(BaseModel):
    cloud_provider: str
    bucket_name: str
    interval_minutes: Optional[int] = None  # For scheduling backups at intervals
    backup_limit: Optional[int] = Field(5, description="Number of backups to retain")

    def save_backup_config(self, model_bazaar_dir):
        config_path = os.path.join(model_bazaar_dir, "backup_config.json")
        os.makedirs(os.path.dirname(config_path), exist_ok=True)

        # Use dict() and json.dump to save the config to a file
        with open(config_path, "w") as file:
            json.dump(self.dict(), file, indent=4)

        return config_path


# S3 specific configuration
class S3BackupConfig(BackupConfig):
    aws_access_key: Optional[str] = Field(None, description="AWS Access Key")
    aws_secret_access_key: Optional[str] = Field(None, description="AWS Secret Key")


# Azure specific configuration
class AzureBackupConfig(BackupConfig):
    azure_account_name: str = Field(..., description="Azure Storage Account Name")
    azure_account_key: str = Field(..., description="Azure Storage Account Key")


# GCP specific configuration
class GCPBackupConfig(BackupConfig):
    gcp_credentials_file_path: str = Field(
        ..., description="GCP Credentials JSON File Path"
    )


def get_cloud_config_class(cloud_provider: str) -> Type[BackupConfig]:
    provider_classes = {
        "s3": S3BackupConfig,
        "azure": AzureBackupConfig,
        "gcp": GCPBackupConfig,
    }
    if cloud_provider in provider_classes:
        return provider_classes[cloud_provider]
    raise HTTPException(
        status_code=status.HTTP_400_BAD_REQUEST,
        detail=f"Unsupported cloud provider: {cloud_provider}",
    )


RECOVERY_SNAPSHOT_ID = "recovery-snapshot"


@recovery_router.post("/backup", dependencies=[Depends(verify_access_token)])
def backup(config: dict):
    local_dir = os.getenv("SHARE_DIR")
    if not local_dir:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="SHARE_DIR environment variable is not set.",
        )

    db_uri = os.getenv("DATABASE_URI")
    if not db_uri:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="DATABASE_URI environment variable is not set.",
        )

    cloud_provider = config.get("cloud_provider")
    if not cloud_provider:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="cloud_provider field is required.",
        )

    # Dynamically select the correct Pydantic model for the cloud provider
    ConfigClass = get_cloud_config_class(cloud_provider)

    # Convert the incoming dictionary to the correct Pydantic model
    config_object = ConfigClass(**config)

    config_file_path = config_object.save_backup_config(local_dir)

    nomad_endpoint = os.getenv("NOMAD_ENDPOINT")
    if nomad_job_exists(RECOVERY_SNAPSHOT_ID, nomad_endpoint):
        delete_nomad_job(RECOVERY_SNAPSHOT_ID, nomad_endpoint)
    cwd = Path(os.getcwd())
    platform = get_platform()
    try:
        submit_nomad_job(
            nomad_endpoint=nomad_endpoint,
            filepath=str(
                cwd / "backend" / "nomad_jobs" / "recovery_snapshot_job.hcl.j2"
            ),
            platform=platform,
            tag=os.getenv("TAG"),
            registry=os.getenv("DOCKER_REGISTRY"),
            docker_username=os.getenv("DOCKER_USERNAME"),
            docker_password=os.getenv("DOCKER_PASSWORD"),
            image_name=os.getenv("RECOVERY_SNAPSHOT_IMAGE_NAME"),
            model_bazaar_endpoint=os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT"),
            python_path=get_python_path(),
            recovery_snapshot_script=str(
                get_root_absolute_path() / "recovery_snapshot_job/run.py"
            ),
            config_path=config_file_path,
            share_dir=local_dir,
            db_uri=db_uri,
        )
    except Exception as err:
        traceback.print_exc()
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, message=str(err)
        )

    return response(
        status_code=status.HTTP_200_OK,
        message="Successfully submitted recovery snapshot job.",
    )
