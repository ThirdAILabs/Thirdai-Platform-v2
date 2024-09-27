import os
from abc import ABC, abstractmethod
from typing import List

from backend.train_config import FileInfo, FileLocation
from fastapi import HTTPException, UploadFile, status


def download_local_file(file_info: FileInfo, upload_file: UploadFile, dest_dir: str):
    assert os.path.basename(file_info.path) == upload_file.filename
    destination_path = os.path.join(dest_dir, upload_file.filename)
    os.makedirs(os.path.dirname(destination_path), exist_ok=True)
    with open(destination_path, "wb") as f:
        f.write(upload_file.file.read())
    upload_file.file.close()
    return destination_path


def download_local_files(
    files: List[UploadFile], file_infos: List[FileInfo], dest_dir: str
) -> List[FileInfo]:
    filename_to_file = {file.filename: file for file in files}

    os.makedirs(dest_dir, exist_ok=True)

    all_files = []
    for file_info in file_infos:
        if file_info.location == FileLocation.local:
            try:
                local_path = download_local_file(
                    file_info=file_info,
                    upload_file=filename_to_file[os.path.basename(file_info.path)],
                    dest_dir=dest_dir,
                )
            except Exception as error:
                raise ValueError(
                    f"Error processing file '{file_info.path}' from '{file_info.location}': {error}"
                )
            all_files.append(
                FileInfo(
                    path=local_path,
                    location=file_info.location,
                    doc_id=file_info.doc_id,
                    options=file_info.options,
                    metadata=file_info.metadata,
                )
            )
        else:
            all_files.append(file_info)

    return all_files


class CloudStorageHandler(ABC):
    """
    Interface for Cloud Storage Handlers.
    All cloud storage handlers must implement these methods.
    """

    @abstractmethod
    def create_bucket_if_not_exists(self, bucket_name: str):
        pass

    @abstractmethod
    def upload_file(self, source_path: str, bucket_name: str, dest_path: str):
        pass

    @abstractmethod
    def upload_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        pass

    @abstractmethod
    def download_file(self, bucket_name: str, source_path: str, dest_path: str):
        pass

    @abstractmethod
    def download_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        pass

    @abstractmethod
    def list_files(self, bucket_name: str, source_path: str):
        pass


class S3StorageHandler(CloudStorageHandler):
    """
    S3 storage handler implementation.
    """

    def __init__(self):
        self.s3_client = self.create_s3_client()

    def create_s3_client(self):
        import boto3
        from botocore import UNSIGNED
        from botocore.client import Config

        aws_access_key = os.getenv("AWS_ACCESS_KEY")
        aws_secret_access_key = os.getenv("AWS_ACCESS_SECRET")

        if not aws_access_key or not aws_secret_access_key:
            config = Config(
                signature_version=UNSIGNED,
                retries={"max_attempts": 10, "mode": "standard"},
                connect_timeout=5,
                read_timeout=60,
            )
            s3_client = boto3.client("s3", config=config)
        else:
            config = Config(
                retries={"max_attempts": 10, "mode": "standard"},
                connect_timeout=5,
                read_timeout=60,
            )
            s3_client = boto3.client(
                "s3",
                aws_access_key_id=aws_access_key,
                aws_secret_access_key=aws_secret_access_key,
                config=config,
            )
        return s3_client

    def create_bucket_if_not_exists(self, bucket_name: str):
        import boto3
        from botocore.exceptions import ClientError

        try:
            self.s3_client.head_bucket(Bucket=bucket_name)
            print(f"Bucket {bucket_name} already exists.")
        except ClientError as e:
            error_code = int(e.response["Error"]["Code"])
            if error_code == 404:
                try:
                    self.s3_client.create_bucket(
                        Bucket=bucket_name,
                        CreateBucketConfiguration={
                            "LocationConstraint": (
                                boto3.session.Session().region_name
                                if boto3.session.Session().region_name
                                else "us-east-1"
                            )
                        },
                    )
                    print(f"Bucket {bucket_name} created successfully.")
                except ClientError as e:
                    if e.response["Error"]["Code"] == "BucketAlreadyExists":
                        print(f"Bucket {bucket_name} already exists globally.")
                    elif e.response["Error"]["Code"] == "AccessDenied":
                        raise HTTPException(
                            status_code=status.HTTP_403_FORBIDDEN,
                            detail=f"Access denied to create bucket {bucket_name}. Error: {str(e)}",
                        )
                    else:
                        raise HTTPException(
                            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                            detail=f"Failed to create bucket {bucket_name}. Error: {str(e)}",
                        )
            else:
                raise HTTPException(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    detail=f"Error checking bucket {bucket_name}. Error: {str(e)}",
                )
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to access bucket {bucket_name}. Error: {str(e)}",
            )

    def upload_file(self, source_path: str, bucket_name: str, dest_path: str):
        try:
            self.s3_client.upload_file(source_path, bucket_name, dest_path)
            print(f"Uploaded {source_path} to {bucket_name}/{dest_path}.")
        except Exception as e:
            print(f"Failed to upload {source_path}. Error: {str(e)}")

    def upload_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        for root, _, files in os.walk(source_dir):
            for file in files:
                local_path = os.path.join(root, file)
                relative_path = os.path.relpath(local_path, source_dir)
                s3_path = os.path.join(dest_dir, relative_path)
                self.upload_file(local_path, bucket_name, s3_path)

    def download_file(self, bucket_name: str, source_path: str, dest_path: str):
        try:
            self.s3_client.download_file(bucket_name, source_path, dest_path)
            print(f"Downloaded {source_path} to {dest_path}.")
        except Exception as e:
            print(f"Failed to download {source_path}. Error: {str(e)}")

    def download_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        s3_files = self.list_files(bucket_name=bucket_name, prefix=source_dir)

        os.makedirs(dest_dir, exist_ok=True)

        for s3_file in s3_files:
            object_key = s3_file.replace(f"s3://{bucket_name}/", "")

            # Define the relative path for the local destination
            relative_path = object_key[len(source_dir) :].lstrip("/")
            dest_path = os.path.join(dest_dir, relative_path)

            # Ensure the directory structure exists locally
            os.makedirs(os.path.dirname(dest_path), exist_ok=True)
            self.download_file(bucket_name, object_key, dest_path)

    def list_files(self, bucket_name: str, source_path: str):
        paginator = self.s3_client.get_paginator("list_objects_v2")
        pages = paginator.paginate(Bucket=bucket_name, Prefix=source_path)
        file_keys = [
            f"s3://{bucket_name}/{obj['Key']}"
            for page in pages
            if "Contents" in page
            for obj in page["Contents"]
            if obj["Key"][-1] != "/"
        ]
        return file_keys


class GCPStorageHandler(CloudStorageHandler):
    """
    GCP storage handler implementation.
    """

    def create_bucket_if_not_exists(self, bucket_name: str):
        # TODO: Implement using GCP SDK (Google Cloud Storage client)
        pass

    def upload_file(self, source_path: str, bucket_name: str, dest_path: str):
        # TODO: Implement using GCP SDK
        pass

    def upload_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        # TODO: Implement using GCP SDK
        pass

    def download_file(self, bucket_name: str, source_path: str, dest_path: str):
        # TODO: Implement using GCP SDK
        pass

    def download_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        # TODO: Implement using GCP SDK
        pass

    def list_files(self, bucket_name: str, source_path: str):
        # TODO: Implement using GCP SDK
        pass


class AzureStorageHandler(CloudStorageHandler):
    """
    Azure storage handler implementation.
    """

    def create_bucket_if_not_exists(self, bucket_name: str):
        # TODO: Implement using Azure Blob Storage SDK
        pass

    def upload_file(self, source_path: str, bucket_name: str, dest_path: str):
        # TODO: Implement using Azure SDK
        pass

    def upload_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        # TODO: Implement using Azure SDK
        pass

    def download_file(self, bucket_name: str, source_path: str, dest_path: str):
        # TODO: Implement using Azure SDK
        pass

    def download_folder(self, bucket_name: str, source_dir: str, dest_dir: str):
        # TODO: Implement using Azure SDK
        pass

    def list_files(self, bucket_name: str, source_path: str):
        # TODO: Implement using Azure SDK
        pass
