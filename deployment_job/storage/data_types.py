from __future__ import annotations
from typing import Optional, List, Dict, Union, ClassVar
from uuid import uuid4
from pydantic import BaseModel, Field


"""
These datatypes are helper objects for storing data into a persistent storage 
without the hassle of saving and loading the entire object on disk.

Example Usescase:

NER :
    TokenClassificationSample : Used for storing user provided training data or data generated using our datagen service.
    TokenClassificationFeedback : Used for storing feedback given by a user about a sample. 
    TagMetaData : Used for storing what tags are present in the pipeline along with their status (trained/untrained)
"""


class LabelEntity(BaseModel):
    name: str
    status: str = "untrained"
    examples: Optional[List[str]] = None
    description: Optional[str] = None
    sample: Optional[str] = None


class LabelEntityList(BaseModel):
    tags: List[LabelEntity]


class SerializableModel(BaseModel):
    def serialize(self) -> str:
        return self.model_dump_json()

    @classmethod
    def deserialize(cls, repr: str) -> "SerializableModel":
        return cls.model_validate_json(repr)


# Text Classification Sample
class TextClassificationSample(SerializableModel):
    datatype: ClassVar[str] = "text_classification"
    text: str
    label: str


# Token Classification Sample
class TokenClassificationSample(SerializableModel):
    datatype: ClassVar[str] = "token_classification"
    tokens: List[str]
    tags: List[str]


# DataSample containing the actual sample object and metadata
class DataSample(BaseModel):
    name: str
    sample: Union[TextClassificationSample, TokenClassificationSample]
    unique_id: str = Field(default_factory=lambda: str(uuid4()))
    user_provided: bool = False

    def serialize_sample(self) -> str:
        return self.sample.serialize()

    @staticmethod
    def deserialize(
        type: str,
        unique_id: str,
        name: str,
        serialized_sample: str,
        user_provided: bool,
    ) -> "DataSample":
        # Deserialize the sample based on its type
        if type == TextClassificationSample.datatype:
            sample = TextClassificationSample.deserialize(serialized_sample)
        elif type == TokenClassificationSample.datatype:
            sample = TokenClassificationSample.deserialize(serialized_sample)
        else:
            raise ValueError(f"Unknown sample type: {type}")

        return DataSample(
            name=name, sample=sample, unique_id=unique_id, user_provided=user_provided
        )

    @property
    def datatype(self):
        return self.sample.datatype


class TokenClassificationFeedBack(SerializableModel):
    datatype: ClassVar[str] = "token_classification"
    delimiter: str
    index_to_label: Dict[int, str]


class UserFeedBack(BaseModel):
    name: str
    sample_uuid: str
    feedback: TokenClassificationFeedBack

    def serialize_feedback(self) -> str:
        return self.feedback.serialize()

    @staticmethod
    def deserialize(
        type: str, name: str, sample_uuid: str, serialized_feedback: str
    ) -> "UserFeedBack":
        if type == TokenClassificationFeedBack.datatype:
            feedback = TokenClassificationFeedBack.deserialize(serialized_feedback)
        else:
            raise ValueError(f"Unknown feedback type: {type}")

        return UserFeedBack(name=name, sample_uuid=sample_uuid, feedback=feedback)

    @property
    def datatype(self):
        return self.feedback.datatype


class TagMetadata(SerializableModel):
    datatype: ClassVar[str] = "token_classification_tags"
    tag_and_status: Dict[str, LabelEntity] = Field(default_factory=dict)

    def update_tag_status(self, tag: str, status: str):
        if tag in self.tag_and_status:
            self.tag_and_status[tag].status = status
        else:
            raise ValueError(f"Tag {tag} not found")

    def add_tag(self, tag: LabelEntity):
        if tag in self.tag_and_status:
            raise Exception(f"Tag {tag.name} is already present in the Tag List")

        self.tag_and_status[tag.name] = tag


class ModelMetadata(BaseModel):
    name: str
    metadata: TagMetadata

    def serialize_metadata(self) -> str:
        return self.metadata.serialize()

    def deserialize(type: str, name: str, serialized_metadata: str):
        if type == TagMetadata.datatype:
            metadata = TagMetadata.deserialize(serialized_metadata)
        else:
            raise ValueError(f"Unknown metadata type: {type}")

        return ModelMetadata(name=name, metadata=metadata)

    @property
    def datatype(self):
        return self.metadata.datatype
