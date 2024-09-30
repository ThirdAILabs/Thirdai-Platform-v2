import datetime
import os
import subprocess
from typing import Optional, Type

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
    backup_limit: Optional[int] = Field(5, description="Number of backups to retain")


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
    if cloud_provider == "s3":
        return S3BackupConfig
    elif cloud_provider == "azure":
        return AzureBackupConfig
    elif cloud_provider == "gcp":
        return GCPBackupConfig
    else:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Unsupported cloud provider: {cloud_provider}",
        )



def get_cloud_storage_handler(config: BackupConfig):
    if isinstance(config, S3BackupConfig):
        return S3StorageHandler(
            aws_access_key=config.aws_access_key,
            aws_secret_access_key=config.aws_secret_access_key,
        )
    elif isinstance(config, AzureBackupConfig):
        return AzureStorageHandler(
            config.azure_account_name, config.azure_account_key
        )
    elif isinstance(config, GCPBackupConfig):
        return GCPStorageHandler(config.gcp_credentials_file_path)
    else:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Unsupported cloud provider: {config.cloud_provider}",
        )



def perform_backup(
    cloud_handler: CloudStorageHandler,
    bucket_name: str,
    db_uri: str,
    local_dir: str,
    backup_limit: int,
):
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")

    dump_file_path = os.path.join(local_dir, f"db_backup_{timestamp}.sql")
    dump_postgres_db_to_file(db_uri, dump_file_path)

    cloud_handler.upload_file(dump_file_path, bucket_name, f"db_backup_{timestamp}.sql")
    cloud_handler.upload_folder(bucket_name, local_dir, f"backups_{timestamp}")

    manage_backup_limit(cloud_handler, bucket_name, backup_limit)

    print(f"Backup to {cloud_handler.__class__.__name__} completed successfully.")


def manage_backup_limit(
    cloud_handler: CloudStorageHandler, bucket_name: str, backup_limit: int
):
    # List the files in the bucket, assuming timestamped backups
    all_backups = cloud_handler.list_files(bucket_name, "backups_")

    # Sort the backups by timestamp (assumed to be part of the filename)
    sorted_backups = sorted(all_backups, key=lambda x: x.split("_")[-1], reverse=True)

    # Keep only the last `backup_limit` backups, delete the rest
    if len(sorted_backups) > backup_limit:
        for backup in sorted_backups[backup_limit:]:
            cloud_handler.delete_object(bucket_name, backup)
            print(f"Deleted old backup: {backup}")


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
    interval_minutes: int, cloud_handler, bucket_name, db_uri, local_dir, backup_limit
):
    scheduler.add_job(
        perform_backup,
        trigger=IntervalTrigger(minutes=interval_minutes),
        args=[cloud_handler, bucket_name, db_uri, local_dir, backup_limit],
        id="backup_job",
        replace_existing=True,  # This replaces any previous job with the same ID
    )


def remove_backup_task():
    if scheduler.get_job("backup_job"):
        scheduler.remove_job("backup_job")
        print("Previous backup job stopped.")


@recovery_router.post("/backup")
def backup_to_s3(config: dict):
    print(config, "at the entry point")
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
        
    cloud_provider = config.get("cloud_provider")
    if not cloud_provider:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="cloud_provider field is required."
        )
        
    # Dynamically select the correct Pydantic model for the cloud provider
    ConfigClass = get_cloud_config_class(cloud_provider)
    
    # Convert the incoming dictionary to the correct Pydantic model
    config_object = ConfigClass(**config)

    cloud_handler = get_cloud_storage_handler(config_object)

    # Ensure the bucket exists
    cloud_handler.create_bucket_if_not_exists(config_object.bucket_name)

    # If interval_minutes is set, schedule a recurring backup
    if config_object.interval_minutes:
        # Stop the previous scheduled backup task if any
        remove_backup_task()
        # Schedule the new task
        schedule_backup_task(
            config_object.interval_minutes,
            cloud_handler,
            config_object.bucket_name,
            db_uri,
            local_dir,
            config_object.backup_limit,
        )
        return response(
            status_code=status.HTTP_200_OK,
            message=f"Scheduled backup to {config_object.cloud_provider} every {config_object.interval_minutes} minutes, keeping the last {config_object.backup_limit} backups.",
        )

    # Perform a one-time backup
    perform_backup(
        cloud_handler, config_object.bucket_name, db_uri, local_dir, config_object.backup_limit
    )

    return response(
        status_code=status.HTTP_200_OK,
        message=f"One-time backup to {config_object.cloud_provider} completed successfully, keeping the last {config_object.backup_limit} backups.",
    )
