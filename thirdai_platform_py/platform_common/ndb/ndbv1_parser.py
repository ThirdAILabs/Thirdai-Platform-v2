import os
from pathlib import Path
from typing import Any, Dict, Optional

from fastapi import Response
from platform_common.file_handler import FileInfo, FileLocation, download_file
from thirdai import neural_db as ndb


def convert_to_ndb_file(
    file: str, metadata: Optional[Dict[str, Any]], options: Dict[str, Any]
) -> ndb.Document:
    """
    Convert a file to an NDB file type based on its extension.
    """
    filename, ext = os.path.splitext(file)
    ext = ext.lower()
    if ext == ".pdf":
        return ndb.PDF(file, metadata=metadata, save_extra_info=False, version="v1")
    elif ext == ".docx":
        return ndb.DOCX(file, metadata=metadata)
    elif ext == ".html":
        base_filename = os.path.basename(filename)

        with open(file, "r", encoding="utf-8") as f:
            html_content = f.read()

        dummy_response = Response()
        dummy_response.status_code = 200
        dummy_response._content = html_content.encode("utf-8")

        return ndb.URL(
            base_filename, dummy_response, metadata=metadata, save_extra_info=False
        )
    elif ext == ".csv":
        return ndb.CSV(
            file,
            id_column=options.get("csv_id_column", None),
            strong_columns=options.get("csv_strong_columns", None),
            weak_columns=options.get("csv_weak_columns", None),
            reference_columns=options.get("csv_reference_columns", None),
            metadata=metadata,
            save_extra_info=False,
        )
    else:
        raise TypeError(f"{ext} Document type isn't supported yet.")


def parse_doc(doc: FileInfo, tmp_dir: str) -> ndb.Document:
    """
    Process a file, downloading it from S3, Azure, or GCP if necessary,
    and convert it to an NDB file.
    """
    # Download the file if it's stored in cloud
    if doc.location in {FileLocation.s3, FileLocation.azure, FileLocation.gcp}:
        local_file_path = download_file(doc, tmp_dir)
        if not local_file_path:
            return f"There was an error downloading the file from {doc.location}. {doc.path}"
    else:
        local_file_path = doc.path

    # Convert the downloaded or local file into an NDB file
    ndb_file = convert_to_ndb_file(
        local_file_path, metadata=doc.metadata, options=doc.options
    )

    # Handle cleanup and adjust file paths for cloud storage
    if doc.location == FileLocation.s3:
        bucket_name, prefix = doc.parse_s3_url()
        ndb_file.path = Path(f"/{bucket_name}.s3.amazonaws.com/{prefix}")
    elif doc.location == FileLocation.azure:
        account_name = os.getenv("AZURE_ACCOUNT_NAME")
        container_name, blob_name = doc.parse_azure_url()
        ndb_file.path = Path(
            f"/{account_name}.blob.core.windows.net/{container_name}/{blob_name}"
        )
    elif doc.location == FileLocation.gcp:
        bucket_name, blob_name = doc.parse_gcp_url()
        ndb_file.path = Path(f"/storage.googleapis.com/{bucket_name}/{blob_name}")

    # Remove the local file if it was downloaded from cloud storage
    if doc.location in {FileLocation.s3, FileLocation.azure, FileLocation.gcp}:
        os.remove(local_file_path)

    return ndb_file
