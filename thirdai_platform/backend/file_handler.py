import json
import os
from abc import ABC, abstractmethod
from enum import Enum
from typing import Dict, List, Optional

import boto3
from fastapi import UploadFile
from pydantic import BaseModel, validator


class FileType(str, Enum):
    unsupervised = "unsupervised"
    supervised = "supervised"
    test = "test"


class FileLocation(str, Enum):
    local = "local"
    nfs = "nfs"
    s3 = "s3"


# BaseModel with validations
class BasicFileDetails(BaseModel):
    mode: FileType
    location: FileLocation
    is_folder: Optional[bool] = False

    @validator("location")
    def check_location(cls, v):
        if v not in FileLocation.__members__.values():
            raise ValueError(
                f"Invalid location value. Supported locations are {list(FileLocation)}"
            )
        return v

    @validator("is_folder", always=True)
    def check_is_folder(cls, v, values):
        if v and values.get("location") == FileLocation.local:
            raise ValueError("is_folder can only be True for nfs and s3 locations.")
        return v

    def validate_csv_extension(self, filename: str):
        if self.mode in {FileType.supervised, FileType.test}:
            _, ext = os.path.splitext(filename)
            if ext != ".csv":
                raise ValueError(
                    f"{filename} file has to be a csv file but given {ext} file."
                )
        return True

    def save_relations(self, supervised_filenames, data_id, source_ids):
        pass  # Default implementation does nothing


class NDBFileDetails(BasicFileDetails):
    source_id: Optional[str] = None
    metadata: Optional[Dict[str, str]] = None

    @validator("source_id", always=True)
    def check_source_id(cls, v, values):
        if values.get("mode") == FileType.supervised and not v:
            raise ValueError("source_id is required for ndb supervised files.")
        if values.get("mode") != FileType.supervised and v:
            raise ValueError("source_id is only allowed for supervised files.")
        if values.get("source_id") and v is not None:
            raise ValueError("source_id should not be provided when is_folder is True")
        return v

    @validator("metadata", always=True)
    def check_metadata(cls, v, values):
        if values.get("is_folder") and v is not None:
            raise ValueError("metadata should not be provided when is_folder is True")
        if values.get("mode") == FileType.supervised and v is not None:
            raise ValueError(
                "metadata is not required for supervised files, it is only for unsupervised files."
            )
        return v

    @validator("is_folder", always=True)
    def check_is_folder(cls, v, values):
        if v and values.get("mode") == FileType.supervised:
            raise ValueError("is_folder cannot be True for ndb supervised files.")
        return v

    def save_relations(self, supervised_filenames, data_id, source_ids):
        if len(supervised_filenames) != len(source_ids):
            raise ValueError("Source ids have not been given for all supervised files.")

        relations_dict_list = [
            {
                "supervised_file": os.path.basename(file_name),
                "source_id": source_ids[i],
            }
            for i, file_name in enumerate(supervised_filenames)
        ]

        destination_path = os.path.join(
            os.getenv("LOCAL_TEST_DIR", "/model_bazaar"),
            "data",
            str(data_id),
            "relations.json",
        )

        os.makedirs(os.path.dirname(destination_path), exist_ok=True)

        with open(destination_path, "w") as file:
            json.dump(relations_dict_list, file, indent=4)


class UDTFileDetails(BasicFileDetails):
    @validator("mode")
    def check_mode(cls, v):
        if v == FileType.unsupervised:
            raise ValueError("UDT files cannot be in 'unsupervised' mode")
        return v


class NDBFileDetailsList(BaseModel):
    file_details: List[NDBFileDetails]

    @validator("file_details")
    def check_file_counts(cls, v):
        test_count = sum(1 for file in v if file.mode == FileType.test)
        unsupervised_count = sum(1 for file in v if file.mode == FileType.unsupervised)
        unsupervised_metadata_count = sum(
            1
            for file in v
            if file.mode == FileType.unsupervised and file.metadata is not None
        )

        if test_count > 1:
            raise ValueError("Currently supports a single test file")

        if 0 < unsupervised_metadata_count < unsupervised_count:
            raise ValueError(
                "Either all unsupervised files must have metadata or none should have metadata."
            )

        return v


class UDTFileDetailsList(BaseModel):
    file_details: List[UDTFileDetails]

    @validator("file_details")
    def check_file_counts(cls, v):
        test_count = sum(1 for file in v if file.mode == FileType.test)

        if test_count > 1:
            raise ValueError("Currently supports a single test file")

        return v


def get_files(files: List[UploadFile], data_id, files_info: List[BasicFileDetails]):
    filenames = []
    supervised_filenames = []
    source_ids = []

    for i, file in enumerate(files):
        file_info = files_info[i]
        destination_dir = os.path.join(
            os.getenv("LOCAL_TEST_DIR", "/model_bazaar"),
            "data",
            str(data_id),
            file_info.mode,
        )
        os.makedirs(destination_dir, exist_ok=True)

        handler = StorageHandlerFactory.get_handler(file_info.location)()

        if file_info.mode == FileType.supervised:
            supervised_filenames.append(file.filename)
            if isinstance(file_info, NDBFileDetails):
                source_ids.append(file_info.source_id)

        try:
            file_keys = handler.process_files(file_info, file, destination_dir)
            filenames.extend(file_keys)
        except Exception as error:
            return f"Error processing file from {file_info.location}: {error}"

        if hasattr(file_info, "metadata") and file_info.metadata:
            metadata_file_path = f"{os.path.splitext(file.filename)[0]}_metadata.json"
            with open(metadata_file_path, "w") as json_file:
                json.dump(file_info.metadata, json_file)

    if supervised_filenames and source_ids:
        # Call save_relations method on the specific class
        for file_info in files_info:
            if isinstance(file_info, NDBFileDetails):
                file_info.save_relations(supervised_filenames, data_id, source_ids)

    return filenames


# Base storage handler
class StorageHandler(ABC):
    @abstractmethod
    def process_files(
        self, file_info: BasicFileDetails, file: UploadFile, destination_dir: str
    ):
        pass

    @abstractmethod
    def validate_file(self, file_info: BasicFileDetails, filename: str):
        pass


# Local storage handler
class LocalStorageHandler(StorageHandler):
    def process_files(
        self, file_info: BasicFileDetails, file: UploadFile, destination_dir: str
    ):
        destination_path = os.path.join(destination_dir, file.filename)
        os.makedirs(os.path.dirname(destination_path), exist_ok=True)
        with open(destination_path, "wb") as f:
            f.write(file.file.read())
        file.file.close()
        self.validate_file(file_info, destination_path)
        return [destination_path]

    def validate_file(self, file_info: BasicFileDetails, filename: str):
        file_info.validate_csv_extension(filename)


# NFS storage handler
class NFSStorageHandler(StorageHandler):
    def process_files(
        self, file_info: BasicFileDetails, file: UploadFile, destination_dir: str
    ):
        file_keys = []
        nfs_file_path = os.path.join(destination_dir, "nfs_files.txt")
        os.makedirs(os.path.dirname(nfs_file_path), exist_ok=True)
        if file_info.is_folder:
            for root, _, files_in_dir in os.walk(file.filename):
                for filename in files_in_dir:
                    src_file_path = os.path.join(root, filename)
                    with open(nfs_file_path, "a") as nfs_file:
                        nfs_file.write(src_file_path + "\n")
                    file_keys.append(src_file_path)
                    self.validate_file(file_info, src_file_path)
        else:
            with open(nfs_file_path, "a") as nfs_file:
                nfs_file.write(file.filename + "\n")
            file_keys.append(file.filename)
            self.validate_file(file_info, file.filename)
        return file_keys

    def validate_file(self, file_info: BasicFileDetails, filename: str):
        file_info.validate_csv_extension(filename)


# S3 storage handler
class S3StorageHandler(StorageHandler):
    def create_s3_client(self):
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
            s3_client = boto3.client(
                "s3",
                config=config,
            )
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

    def process_files(
        self, file_info: BasicFileDetails, file: UploadFile, destination_dir: str
    ):
        s3_file_path = os.path.join(destination_dir, "s3_files.txt")
        os.makedirs(os.path.dirname(s3_file_path), exist_ok=True)
        s3_files = self.list_s3_files(file.filename)
        for s3_file in s3_files:
            with open(s3_file_path, "a") as s3_file_local:
                s3_file_local.write(s3_file + "\n")
            self.validate_file(file_info, s3_file)
        return s3_files

    def list_s3_files(self, filename):
        s3 = self.create_s3_client()
        bucket_name, prefix = filename.replace("s3://", "").split("/", 1)
        paginator = s3.get_paginator("list_objects_v2")
        pages = paginator.paginate(Bucket=bucket_name, Prefix=prefix)
        file_keys = []
        for page in pages:
            if "Contents" in page:
                for obj in page["Contents"]:
                    file_keys.append(f"s3://{bucket_name}/{obj['Key']}")
        return file_keys

    def validate_file(self, file_info: BasicFileDetails, filename: str):
        file_info.validate_csv_extension(filename)


# Factory to get the correct handler
class StorageHandlerFactory:
    handlers = {
        FileLocation.local: LocalStorageHandler,
        FileLocation.nfs: NFSStorageHandler,
        FileLocation.s3: S3StorageHandler,
    }

    @classmethod
    def get_handler(cls, location):
        handler_class = cls.handlers.get(location)
        if not handler_class:
            raise ValueError(f"No handler found for location: {location}")
        return handler_class
