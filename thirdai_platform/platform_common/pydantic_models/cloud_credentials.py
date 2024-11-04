from typing import Optional

from pydantic import BaseModel, Field


class AWSCredentials(BaseModel):
    access_key: Optional[str] = Field(None, env="AWS_ACCESS_KEY")
    secret_key: Optional[str] = Field(None, env="AWS_ACCESS_SECRET")
    region_name: Optional[str] = Field(None, env="AWS_REGION_NAME")


class AzureCredentials(BaseModel):
    account_name: Optional[str] = Field(None, env="AZURE_ACCOUNT_NAME")
    account_key: Optional[str] = Field(None, env="AZURE_ACCOUNT_KEY")


class GCPCredentials(BaseModel):
    credentials_file: Optional[str] = Field(None, env="GCP_CREDENTIALS_FILE")


# Main training configuration model
class CloudCredentials(BaseModel):
    aws: Optional[AWSCredentials] = None
    azure: Optional[AzureCredentials] = None
    gcp: Optional[GCPCredentials] = None
