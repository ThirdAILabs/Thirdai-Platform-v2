import datetime
import json
import logging
import os
import shutil
import subprocess

from apscheduler.schedulers.blocking import BlockingScheduler
from apscheduler.triggers.date import DateTrigger
from apscheduler.triggers.interval import IntervalTrigger
from file_handler import (
    AzureStorageHandler,
    CloudStorageHandler,
    GCPStorageHandler,
    S3StorageHandler,
)

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)

# Global scheduler instance
scheduler = BlockingScheduler()


def load_config(config_file):
    with open(config_file, "r") as f:
        return json.load(f)


def get_cloud_storage_handler(config):
    if config["cloud_provider"] == "s3":
        return S3StorageHandler(
            aws_access_key=config["aws_access_key"],
            aws_secret_access_key=config["aws_secret_access_key"],
        )
    elif config["cloud_provider"] == "azure":
        return AzureStorageHandler(
            config["azure_account_name"], config["azure_account_key"]
        )
    elif config["cloud_provider"] == "gcp":
        return GCPStorageHandler(config["gcp_credentials_file_path"])
    else:
        raise ValueError(f"Unsupported cloud provider: {config['cloud_provider']}")


def delete_local_backup(zip_file_path: str, dump_file_path: str):
    try:
        os.remove(f"{zip_file_path}.zip")
        logging.info(f"Deleted local zip file: {zip_file_path}.zip")
        os.remove(dump_file_path)
        logging.info(f"Deleted local DB dump file: {dump_file_path}")
    except Exception as e:
        logging.error(f"Error while deleting local files: {str(e)}")


def create_backup_files(db_uri: str, local_dir: str, timestamp: str):
    dump_file_path = os.path.join(local_dir, f"db_backup_{timestamp}.sql")
    subprocess.run(["pg_dump", db_uri, "-f", dump_file_path], check=True)
    zip_file_path = os.path.join(local_dir, f"backup_{timestamp}")
    shutil.make_archive(zip_file_path, "zip", local_dir)
    return zip_file_path, dump_file_path


def manage_backup_limit(
    cloud_handler: CloudStorageHandler, bucket_name: str, backup_limit: int
):
    all_backups = cloud_handler.list_files(bucket_name, "backup_")
    sorted_backups = sorted(all_backups, key=lambda x: x.split("_")[-1], reverse=True)
    if len(sorted_backups) > backup_limit:
        for backup in sorted_backups[backup_limit:]:
            cloud_handler.delete_path(bucket_name, backup)
            logging.info(f"Deleted old backup: {backup}")


def perform_backup(config_file):
    config = load_config(config_file)
    timestamp = datetime.datetime.now().strftime("%Y%m%d_%H%M%S")

    local_dir = os.getenv("MODEL_BAZAAR_DIR")
    db_uri = os.getenv("DB_URI")
    bucket_name = config["bucket_name"]
    backup_limit = config["backup_limit"]

    zip_file_path, dump_file_path = create_backup_files(db_uri, local_dir, timestamp)

    cloud_handler = get_cloud_storage_handler(config)
    cloud_handler.create_bucket_if_not_exists(bucket_name)
    cloud_handler.upload_file(
        f"{zip_file_path}.zip", bucket_name, f"backup_{timestamp}.zip"
    )

    manage_backup_limit(cloud_handler, bucket_name, backup_limit)
    delete_local_backup(zip_file_path, dump_file_path)

    logging.info(f"Backup completed successfully at {timestamp}.")

    # If this is a one-time backup, shut down the scheduler
    if not config.get("interval_minutes"):
        scheduler.shutdown(wait=False)  # Stops the scheduler after the job completes


def schedule_backup(config_file):
    """Schedule backup based on interval in config or run once."""
    config = load_config(config_file)
    interval_minutes = config.get("interval_minutes")

    if interval_minutes:
        # Schedule recurring backups
        scheduler.add_job(
            perform_backup,
            trigger=IntervalTrigger(minutes=interval_minutes),
            args=[config_file],
            id="backup_job",
            replace_existing=True,
        )
        logging.info(f"Scheduled periodic backup every {interval_minutes} minutes.")
    else:
        # Run one-time backup
        scheduler.add_job(
            perform_backup,
            trigger=DateTrigger(run_date=datetime.datetime.now()),
            args=[config_file],
            id="one_time_backup_job",
            replace_existing=True,
        )
        logging.info("Scheduled one-time backup.")

    # Start the scheduler
    try:
        scheduler.start()
    except Exception as e:
        logging.error(f"Scheduler encountered an error: {e}")
        scheduler.shutdown()


if __name__ == "__main__":
    config_file = os.getenv("CONFIG_PATH")
    if not config_file or not os.path.exists(config_file):
        raise ValueError("Config file not found.")

    try:
        schedule_backup(config_file)
    except Exception as err:
        logging.error(f"Error during backup: {err}")
