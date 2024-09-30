import datetime
import os
import shutil
import subprocess
from typing import Optional, Type

from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.date import DateTrigger
from apscheduler.triggers.interval import IntervalTrigger
from backend.file_handler import (
    AzureStorageHandler,
    CloudStorageHandler,
    GCPStorageHandler,
    S3StorageHandler,
)
from backend.utils import model_bazaar_path, response
from fastapi import APIRouter, HTTPException, status
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


def get_cloud_storage_handler(config: BackupConfig):
    if isinstance(config, S3BackupConfig):
        return S3StorageHandler(
            aws_access_key=config.aws_access_key,
            aws_secret_access_key=config.aws_secret_access_key,
        )
    elif isinstance(config, AzureBackupConfig):
        return AzureStorageHandler(config.azure_account_name, config.azure_account_key)
    elif isinstance(config, GCPBackupConfig):
        return GCPStorageHandler(config.gcp_credentials_file_path)
    else:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Unsupported cloud provider: {config.cloud_provider}",
        )


def delete_local_backup(zip_file_path: str, dump_file_path: str):
    try:
        os.remove(f"{zip_file_path}.zip")
        print(f"Deleted local zip file: {zip_file_path}.zip")
        os.remove(dump_file_path)
        print(f"Deleted local DB dump file: {dump_file_path}")
    except Exception as e:
        print(f"Error while deleting local files: {str(e)}")


def create_backup_files(db_uri: str, local_dir: str, timestamp: str):
    dump_file_path = os.path.join(local_dir, f"db_backup_{timestamp}.sql")
    dump_postgres_db_to_file(db_uri, dump_file_path)

    zip_file_path = os.path.join(local_dir, f"backup_{timestamp}")
    shutil.make_archive(zip_file_path, "zip", local_dir)

    return zip_file_path, dump_file_path


def perform_backup(
    cloud_handler: CloudStorageHandler,
    bucket_name: str,
    db_uri: str,
    local_dir: str,
    backup_limit: int,
):
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")
    zip_file_path, dump_file_path = create_backup_files(db_uri, local_dir, timestamp)

    cloud_handler.upload_file(
        f"{zip_file_path}.zip", bucket_name, f"backup_{timestamp}.zip"
    )

    manage_backup_limit(cloud_handler, bucket_name, backup_limit)
    delete_local_backup(zip_file_path, dump_file_path)

    print(f"Backup to {cloud_handler.__class__.__name__} completed successfully.")


def manage_backup_limit(
    cloud_handler: CloudStorageHandler, bucket_name: str, backup_limit: int
):
    # List the files in the bucket, assuming timestamped backups
    all_backups = cloud_handler.list_files(bucket_name, "backup_")

    # Sort the backups by timestamp (assumed to be part of the filename)
    sorted_backups = sorted(all_backups, key=lambda x: x.split("_")[-1], reverse=True)

    # Keep only the last `backup_limit` backups, delete the rest
    if len(sorted_backups) > backup_limit:
        for backup in sorted_backups[backup_limit:]:
            cloud_handler.delete_path(bucket_name, backup)
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
    cloud_handler: CloudStorageHandler,
    bucket_name: str,
    db_uri: str,
    local_dir: str,
    backup_limit: int,
    interval_minutes: Optional[int] = None,
):
    trigger = (
        IntervalTrigger(minutes=interval_minutes)
        if interval_minutes
        else DateTrigger(run_date=datetime.datetime.now())
    )

    job_id = "backup_job" if interval_minutes else "one_time_backup_job"

    scheduler.add_job(
        perform_backup,
        trigger=trigger,
        args=[cloud_handler, bucket_name, db_uri, local_dir, backup_limit],
        id=job_id,
        replace_existing=True,
    )


def remove_backup_task():
    for job_id in ["backup_job", "one_time_backup_job"]:
        if scheduler.get_job(job_id):
            scheduler.remove_job(job_id)
            print(f"Previous {job_id} stopped.")


@recovery_router.post("/backup")
def backup_to_s3(config: dict):
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
            detail="cloud_provider field is required.",
        )

    # Dynamically select the correct Pydantic model for the cloud provider
    ConfigClass = get_cloud_config_class(cloud_provider)

    # Convert the incoming dictionary to the correct Pydantic model
    config_object = ConfigClass(**config)

    cloud_handler = get_cloud_storage_handler(config_object)

    # Ensure the bucket exists
    cloud_handler.create_bucket_if_not_exists(config_object.bucket_name)

    remove_backup_task()

    # Schedule the backup task (periodic or one-time)
    schedule_backup_task(
        cloud_handler,
        config_object.bucket_name,
        db_uri,
        local_dir,
        config_object.backup_limit,
        config_object.interval_minutes,
    )

    message = (
        f"Scheduled backup to {config_object.cloud_provider} every {config_object.interval_minutes} minutes"
        if config_object.interval_minutes
        else f"One-time backup to {config_object.cloud_provider} scheduled successfully"
    )

    return response(status_code=status.HTTP_200_OK, message=message)
