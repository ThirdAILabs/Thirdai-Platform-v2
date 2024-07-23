from abc import ABC, abstractmethod
import os
import boto3
from botocore import UNSIGNED
from botocore.client import Config
        
class FileHandler:
    def __init__(self):
        pass
    
    @abstractmethod
    def reader(self, **kwargs):
        pass
    
    @abstractmethod
    def writer(self, **kwargs):
        pass
    

class S3FileHandler(FileHandler):
    def __init__(self):
        super().__init__()
        self.s3 = self.create_s3_client()
    
    def create_s3_client(self):
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
                aws_access_key_id=os.getenv("AWS_ACCESS_KEY"),
                aws_secret_access_key=os.getenv("AWS_ACCESS_SECRET"),
                config=config,
            )
        return s3_client

    def list_files_in_bucket(self, bucket_name, prefix):
        paginator = self.s3.get_paginator("list_objects_v2")
        pages = paginator.paginate(Bucket=bucket_name, Prefix=prefix)

        file_keys = []

        for page in pages:
            if "Contents" in page:
                for obj in page["Contents"]:
                    file_keys.append(obj["Key"])

        return file_keys

    def list_s3_files(self, s3_url):
        bucket_name, prefix = s3_url.replace("s3://", "").split("/", 1)
        file_keys = self.list_files_in_s3(bucket_name, prefix)
        s3_urls = [f"s3://{bucket_name}/{key}" for key in file_keys]
        
        return s3_urls
    
    def reader(self, s3_url):
        s3_files = self.list_s3_files(s3_url)
    
    def writer(self, destination_path):
        pass
    
class NFSFileHandler(FileHandler):
    def __init__(self):
        super().__init__()
        
class LocalFileHandler(FileHandler):
    def __init__(self):
        super().__init__()
        
        
