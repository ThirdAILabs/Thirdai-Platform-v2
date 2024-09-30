import os
import subprocess
from typing import Optional

from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.interval import IntervalTrigger
from auth.jwt import verify_access_token
from backend.file_handler import (
    AzureStorageHandler,
    CloudStorageHandler,
    GCPStorageHandler,
    S3StorageHandler,
)
from backend.utils import model_bazaar_path, response
from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel, Field

recovery_router = APIRouter()

scheduler = BackgroundScheduler()
scheduler.start()


# Base configuration model for backup
class BackupConfig(BaseModel):
    cloud_provider: str
    bucket_name: str
    interval_minutes: Optional[int] = None  # For scheduling backups at intervals


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


def get_cloud_storage_handler(config: BackupConfig):
    if config.cloud_provider == "s3":
        s3_config = S3BackupConfig(**config.dict())
        return S3StorageHandler(
            aws_access_key=s3_config.aws_access_key,
            aws_secret_access_key=s3_config.aws_secret_access_key,
        )

    elif config.cloud_provider == "azure":
        azure_config = AzureBackupConfig(**config.dict())
        return AzureStorageHandler(
            azure_config.azure_account_name, azure_config.azure_account_key
        )

    elif config.cloud_provider == "gcp":
        gcp_config = GCPBackupConfig(**config.dict())
        return GCPStorageHandler(gcp_config.gcp_credentials_file_path)

    else:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Unsupported cloud provider: {config.cloud_provider}",
        )


def perform_backup(
    cloud_handler: CloudStorageHandler, bucket_name: str, db_uri: str, local_dir: str
):
    dump_file_path = os.path.join(local_dir, "db_backup.sql")
    dump_postgres_db_to_file(db_uri, dump_file_path)

    cloud_handler.upload_file(dump_file_path, bucket_name, "db_backup.sql")
    cloud_handler.upload_folder(bucket_name, local_dir, "backups")

    print(f"Backup to {cloud_handler.__class__.__name__} completed successfully.")


def dump_postgres_db_to_file(db_uri, dump_file_path):
    try:
        subprocess.run(["pg_dump", db_uri, "-f", dump_file_path], check=True)
        print(f"Database successfully dumped to {dump_file_path}")
    except subprocess.CalledProcessError as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to dump the database: {str(e)}",
        )


def schedule_backup_task(
    interval_minutes: int, cloud_handler, bucket_name, db_uri, local_dir
):
    scheduler.add_job(
        perform_backup,
        trigger=IntervalTrigger(minutes=interval_minutes),
        args=[cloud_handler, bucket_name, db_uri, local_dir],
        id="backup_job",
        replace_existing=True,  # This replaces any previous job with the same ID
    )


def remove_backup_task():
    if scheduler.get_job("backup_job"):
        scheduler.remove_job("backup_job")
        print("Previous backup job stopped.")


@recovery_router.post("/backup", dependencies=[Depends(verify_access_token)])
def backup_to_s3(config: BackupConfig):
    local_dir = model_bazaar_path()
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

    cloud_handler = get_cloud_storage_handler(config)

    # Ensure the bucket exists
    cloud_handler.create_bucket_if_not_exists(config.bucket_name)

    # If interval_minutes is set, schedule a recurring backup
    if config.interval_minutes:
        # Stop the previous scheduled backup task if any
        remove_backup_task()
        # Schedule the new task
        schedule_backup_task(
            config.interval_minutes,
            cloud_handler,
            config.bucket_name,
            db_uri,
            local_dir,
        )
        return response(
            status_code=status.HTTP_200_OK,
            message=f"Scheduled backup to {config.cloud_provider} every {config.interval_minutes} minutes.",
        )

    # Perform a one-time backup
    perform_backup(cloud_handler, config.bucket_name, db_uri, local_dir)

    return response(
        status_code=status.HTTP_200_OK,
        message=f"One-time backup to {config.cloud_provider} completed successfully.",
    )
