import os
from enum import Enum
from typing import Any, Dict, List, Literal, Optional, Union

from platform_common.pydantic_models.llm_config import LLMConfig
from platform_common.thirdai_storage.data_types import (
    LabelEntity,
    TokenClassificationData,
)
from pydantic import BaseModel, Field, model_validator


class ModelType(str, Enum):
    NDB = "ndb"
    NLP_TOKEN = "nlp-token"
    NLP_TEXT = "nlp-text"
    NLP_DOC = "nlp-doc"
    ENTERPRISE_SEARCH = "enterprise-search"
    KNOWLEDGE_EXTRACTION = "ke"


class ModelDataType(str, Enum):
    NDB = "ndb"
    NLP = "nlp"
    UDT_DATAGEN = "udt_datagen"


class FileLocation(str, Enum):
    local = "local"
    nfs = "nfs"
    s3 = "s3"
    azure = "azure"
    gcp = "gcp"


class FileInfo(BaseModel):
    path: str
    location: FileLocation
    source_id: Optional[str] = None
    options: Dict[str, Any] = {}
    metadata: Optional[Dict[str, Union[int, str, float, bool]]] = None

    def ext(self) -> str:
        _, ext = os.path.splitext(self.path)
        return ext

    def parse_s3_url(self):
        """
        Parses an S3 URL and returns the bucket name and key.
        Only works for FileInfo objects where location is S3.

        Returns:
            tuple: (bucket_name, key) where bucket_name is the name of the S3 bucket
                   and key is the path within the bucket.
        """
        if self.location != FileLocation.s3:
            raise ValueError(
                f"Invalid operation. File location is not S3: {self.location}"
            )

        if not self.path.startswith("s3://"):
            raise ValueError(f"Invalid S3 URL: {self.path}")

        # Remove the 's3://' prefix and split the remaining string
        s3_path = self.path[5:]  # Remove 's3://'
        bucket_name, _, key = s3_path.partition("/")

        return bucket_name, key

    def parse_azure_url(self):
        """
        Parses an Azure Blob Storage URL and returns the container name and blob path.
        Only works for FileInfo objects where location is Azure.

        Returns:
            tuple: (container_name, blob_path) where container_name is the name of the
                   Azure container and blob_path is the path to the blob inside the container.
        """
        if self.location != FileLocation.azure:
            raise ValueError(
                f"Invalid operation. File location is not Azure: {self.location}"
            )

        if not self.path.startswith("https://"):
            raise ValueError(f"Invalid Azure Blob URL: {self.path}")

        # Azure Blob URLs follow the pattern: https://<account_name>.blob.core.windows.net/<container>/<blob_path>
        url_parts = self.path.split("/")

        if len(url_parts) < 4:
            raise ValueError(f"Invalid Azure Blob URL structure: {self.path}")

        container_name = url_parts[
            3
        ]  # After 'https://<account_name>.blob.core.windows.net/'
        blob_path = "/".join(url_parts[4:])  # Remaining path as the blob key

        return container_name, blob_path

    def parse_gcp_url(self):
        """
        Parses a GCP Cloud Storage URL and returns the bucket name and key.
        Only works for FileInfo objects where location is GCP.

        Returns:
            tuple: (bucket_name, key) where bucket_name is the name of the GCP bucket
                   and key is the path within the bucket.
        """
        if self.location != FileLocation.gcp:
            raise ValueError(
                f"Invalid operation. File location is not GCP: {self.location}"
            )

        if not self.path.startswith("gs://"):
            raise ValueError(f"Invalid GCP URL: {self.path}")

        gcs_path = self.path[5:]  # Remove 'gs://'
        bucket_name, _, key = gcs_path.partition("/")

        return bucket_name, key


class AutopopulateMetadataInfo(BaseModel):
    attribute_name: str
    description: str


class NDBOptions(BaseModel):
    model_type: Literal[ModelType.NDB] = ModelType.NDB

    on_disk: bool = True
    advanced_search: bool = False
    autopopulate_doc_metadata_fields: Optional[List[AutopopulateMetadataInfo]] = None

    class Config:
        protected_namespaces = ()


class NDBData(BaseModel):
    model_data_type: Literal[ModelDataType.NDB] = ModelDataType.NDB

    unsupervised_files: List[FileInfo] = []
    supervised_files: List[FileInfo] = []

    deletions: List[str] = []

    class Config:
        protected_namespaces = ()

    @model_validator(mode="after")
    def check_nonempty(self):
        if len(self.unsupervised_files) + len(self.supervised_files) == 0:
            raise ValueError(
                "Unsupervised or supervised files must not be non empty for NDB training."
            )
        return self


class NlpTokenOptions(BaseModel):
    model_type: Literal[ModelType.NLP_TOKEN] = ModelType.NLP_TOKEN

    target_labels: List[str]
    source_column: str
    target_column: str
    default_tag: str = "O"


class NlpTextOptions(BaseModel):
    model_type: Literal[ModelType.NLP_TEXT] = ModelType.NLP_TEXT

    text_column: str
    label_column: str
    n_target_classes: int
    delimiter: str = ","


class NlpTrainOptions(BaseModel):
    epochs: int = 1
    learning_rate: float = 0.0001
    batch_size: int = 2048
    max_in_memory_batches: Optional[int] = None
    test_split: Optional[float] = None


class UDTData(BaseModel):
    model_data_type: Literal[ModelDataType.NLP] = ModelDataType.NLP

    supervised_files: List[FileInfo]
    test_files: List[FileInfo] = []

    class Config:
        protected_namespaces = ()


class UDTGeneratedData(BaseModel):
    model_data_type: Literal[ModelDataType.UDT_DATAGEN] = ModelDataType.UDT_DATAGEN
    secret_token: str

    class Config:
        protected_namespaces = ()


class LLMProvider(str, Enum):
    openai = "openai"
    cohere = "cohere"


class NlpTextDatagenOptions(BaseModel):
    model_type: Literal[ModelType.NLP_TEXT] = ModelType.NLP_TEXT

    labels: List[LabelEntity]
    samples_per_label: int
    user_vocab: Optional[List[str]] = None
    user_prompts: Optional[List[str]] = None
    vocab_per_sentence: int = 4

    class Config:
        protected_namespaces = ()

    @model_validator(mode="after")
    def check_target_labels_length(cls, values):
        if len(values.labels) < 2:
            raise ValueError("target_labels must contain at least two labels.")
        return values


class NlpTokenDatagenOptions(BaseModel):
    model_type: Literal[ModelType.NLP_TOKEN] = ModelType.NLP_TOKEN

    tags: List[LabelEntity]
    num_sentences_to_generate: int = 1_000
    num_samples_per_tag: Optional[int] = None

    # example NER samples
    samples: Optional[List[TokenClassificationData]] = None
    templates_per_sample: int = 10

    load_from_storage: bool = False

    class Config:
        protected_namespaces = ()


class DatagenConfig(BaseModel):
    model_id: str
    base_model_id: Optional[str] = None
    model_bazaar_dir: str
    storage_dir: str
    model_bazaar_endpoint: str

    task_prompt: str
    llm_provider: str
    test_size: float = 0.05

    task_options: Union[NlpTokenDatagenOptions, NlpTextDatagenOptions] = Field(
        ..., descriminator="model_type"
    )

    class Config:
        protected_namespaces = ()


class JobOptions(BaseModel):
    allocation_cores: int = Field(1, gt=0)
    allocation_memory: int = Field(6800, gt=500)


class TrainConfig(BaseModel):
    user_id: str
    model_bazaar_dir: str
    license_key: str
    model_bazaar_endpoint: str
    model_id: str
    job_auth_token: str
    base_model_id: Optional[str] = None

    # The model and data fields are separate because the model_options are designed
    # to be passed directly from the parameters supplied to the train endpoint to
    # the train config without requiring additional processing. The data may require
    # some processing, to download files, copy to directories, etc. Thus they are separated
    # so that the model options can be passed through while the data is processed
    # in the train endpoint.
    model_type: ModelType
    model_options: Optional[Union[NDBOptions, NlpTokenOptions, NlpTextOptions]] = Field(
        ..., discriminator="model_type"
    )
    # datagen_options: Optional[DatagenOptions] = None
    job_options: JobOptions

    train_options: Optional[NlpTrainOptions] = NlpTrainOptions()

    data: Union[NDBData, UDTData, UDTGeneratedData] = Field(
        ..., discriminator="model_data_type"
    )

    is_retraining: bool = False

    generative_supervision: bool = False

    llm_config: Optional[LLMConfig] = None

    class Config:
        protected_namespaces = ()

    @model_validator(mode="after")
    def check_model_data_match(self):
        if self.data.model_data_type.value not in self.model_type.value:
            raise ValueError("Model and data fields don't match")
        if self.generative_supervision and self.llm_config is None:
            raise ValueError("Generative supervision requires LLM config")
        return self

    def save_train_config(self):
        config_path = os.path.join(
            self.model_bazaar_dir, "models", str(self.model_id), "train_config.json"
        )
        os.makedirs(os.path.dirname(config_path), exist_ok=True)
        with open(config_path, "w") as file:
            file.write(self.model_dump_json(indent=4))

        return config_path


class QuestionKeywords(BaseModel):
    question: str = Field(..., description="The mandatory question.")
    keywords: Optional[List[str]] = Field(
        default=None, description="Optional keywords for the question."
    )
