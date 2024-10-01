import datetime
import json
import os
import shutil
import subprocess

from file_handler import (
    AzureStorageHandler,
    CloudStorageHandler,
    GCPStorageHandler,
    S3StorageHandler,
)

def load_config(config_file):
    with open(config_file, "r") as f:
        config = json.load(f)
    return config


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
        print(f"Deleted local zip file: {zip_file_path}.zip")
        os.remove(dump_file_path)
        print(f"Deleted local DB dump file: {dump_file_path}")
    except Exception as e:
        print(f"Error while deleting local files: {str(e)}")


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
            print(f"Deleted old backup: {backup}")


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

    print(f"Backup completed successfully.")


if __name__ == "__main__":
    config_file = os.getenv("CONFIG_PATH")
    if not config_file or not os.path.exists(config_file):
        raise ValueError("Config file not found.")
    try:
        perform_backup(config_file)
    except Exception as err:
        print(err)
