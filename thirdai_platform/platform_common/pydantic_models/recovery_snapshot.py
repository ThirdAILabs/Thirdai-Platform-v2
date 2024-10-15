import json
import os
from typing import Optional

from pydantic import BaseModel, Field, root_validator


class BackupConfig(BaseModel):
    """
    Unified backup configuration class that dynamically validates required fields
    based on the cloud provider type (or if no cloud provider is used for local backups).
    """

    cloud_provider: Optional[str] = Field(
        None, description="Cloud provider (s3, azure, gcp, or empty for local)"
    )
    bucket_name: Optional[str] = Field(
        None, description="Cloud bucket name (for cloud providers)"
    )
    interval_minutes: Optional[int] = Field(
        None, description="For scheduling backups at intervals"
    )
    backup_limit: Optional[int] = Field(5, description="Number of backups to retain")

    # Cloud-specific fields (conditionally required based on cloud_provider)
    aws_access_key: Optional[str] = Field(None, description="AWS Access Key (for s3)")
    aws_secret_access_key: Optional[str] = Field(
        None, description="AWS Secret Key (for s3)"
    )
    azure_account_name: Optional[str] = Field(
        None, description="Azure Storage Account Name (for azure)"
    )
    azure_account_key: Optional[str] = Field(
        None, description="Azure Storage Account Key (for azure)"
    )
    gcp_credentials_file_path: Optional[str] = Field(
        None, description="GCP Credentials JSON File Path (for gcp)"
    )

    @root_validator(pre=True)
    def validate_config(cls, values):
        cloud_provider = values.get("cloud_provider")

        if cloud_provider == "s3":
            # For S3, ensure AWS keys are provided
            if not values.get("aws_access_key") or not values.get(
                "aws_secret_access_key"
            ):
                raise ValueError(
                    "AWS credentials (access key and secret key) are required for S3 backups."
                )
            if not values.get("bucket_name"):
                raise ValueError("Bucket name is required for S3 backups.")

        elif cloud_provider == "azure":
            # For Azure, ensure account name and key are provided
            if not values.get("azure_account_name") or not values.get(
                "azure_account_key"
            ):
                raise ValueError(
                    "Azure account name and key are required for Azure backups."
                )
            if not values.get("bucket_name"):
                raise ValueError("Bucket name is required for Azure backups.")

        elif cloud_provider == "gcp":
            # For GCP, ensure the credentials file path is provided
            if not values.get("gcp_credentials_file_path"):
                raise ValueError(
                    "GCP credentials file path is required for GCP backups."
                )
            if not values.get("bucket_name"):
                raise ValueError("Bucket name is required for GCP backups.")

        elif cloud_provider is None:
            # Local backup: no cloud provider means local backup, no extra fields required
            pass

        else:
            raise ValueError(f"Unsupported cloud provider: {cloud_provider}")

        return values

    def save_backup_config(self, model_bazaar_dir):
        config_path = os.path.join(model_bazaar_dir, "backup_config.json")
        os.makedirs(os.path.dirname(config_path), exist_ok=True)

        # Use dict() and json.dump to save the config to a file
        with open(config_path, "w") as file:
            json.dump(self.dict(), file, indent=4)

        return config_path
