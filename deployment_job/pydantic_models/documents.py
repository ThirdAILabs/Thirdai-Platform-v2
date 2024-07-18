from enum import Enum
from typing import Any, List, Literal, Optional, Union

from pydantic import BaseModel, Field, RootModel
from typing_extensions import Annotated


class DocumentLocation(str, Enum):
    LOCAL = "local"
    NFS = "nfs"
    S3 = "s3"


class Document(BaseModel):
    location: DocumentLocation = DocumentLocation.LOCAL
    on_disk: bool = True


class PDF(Document):
    document_type: Literal["PDF"]
    path: str
    metadata: Optional[dict[str, Any]] = None
    version: str = "v1"
    chunk_size: int = 100
    stride: int = 40
    emphasize_first_words: int = 0
    ignore_header_footer: bool = True
    ignore_nonstandard_orientation: bool = True


class CSV(Document):
    document_type: Literal["CSV"]
    path: str
    id_column: Optional[str] = None
    strong_columns: Optional[List[str]] = None
    weak_columns: Optional[List[str]] = None
    reference_columns: Optional[List[str]] = None
    save_extra_info: bool = True
    metadata: Optional[dict[str, Any]] = None
    has_offset: bool = False


class DOCX(Document):
    document_type: Literal["DOCX"]
    path: str
    metadata: Optional[dict[str, Any]] = None


class URL(Document):
    document_type: Literal["URL"]
    url: str
    save_extra_info: bool = True
    title_is_strong: bool = False
    metadata: Optional[dict[str, Any]] = None


class SentenceLevelPDF(Document):
    document_type: Literal["SentenceLevelPDF"]
    path: str
    metadata: Optional[dict[str, Any]] = None


class SentenceLevelDOCX(Document):
    document_type: Literal["SentenceLevelDOCX"]
    path: str
    metadata: Optional[dict[str, Any]] = None


class Unstructured(Document):
    document_type: Literal["Unstructured"]
    path: str
    save_extra_info: bool = True
    metadata: Optional[dict[str, Any]] = None


class InMemoryText(Document):
    document_type: Literal["InMemoryText"]
    name: str
    texts: list[str]
    metadatas: Optional[list[dict[str, Any]]] = None
    global_metadata: Optional[dict[str, Any]] = None


class DocumentList(RootModel):
    root: list[
        Annotated[
            Union[
                PDF,
                CSV,
                DOCX,
                URL,
                SentenceLevelPDF,
                SentenceLevelDOCX,
                Unstructured,
                InMemoryText,
            ],
            Field(discriminator="document_type"),
        ],
    ] = []
