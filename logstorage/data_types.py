from __future__ import annotations

import json
import uuid
import typing
from abc import abstractmethod, abstractproperty, abstractstaticmethod
from dataclasses import dataclass

from sqlalchemy import UUID
import pandas as pd
import re


class DataType:

    @abstractmethod
    def serialize(self) -> str:
        # can encode the data
        pass

    @abstractstaticmethod
    def deserialize(name: str, repr: str):
        # can decode the data
        pass

    @property
    def name(self):
        return self._name


class DataSamples(DataType):
    def __init__(self, unique_id=None):
        self._uuid = unique_id if unique_id is not None else str(uuid.uuid4())

    @property
    def uuid(self):
        return self._uuid

    @abstractstaticmethod
    def deserialize(name: str, repr: str, unique_id: str):
        # can decode the data
        pass


class TextClassificationSample(DataSamples):
    datatype = "text_classification"

    def __init__(self, name: str, text: str, label: str, unique_id: str = None):
        super().__init__(unique_id=unique_id)

        self._text = text
        self._label = label
        self._name = name

    def serialize(self) -> str:
        return json.dumps({"text": self._text, "label": self._label})

    @staticmethod
    def deserialize(name, repr, unique_id) -> str:
        data = json.loads(repr)
        return TextClassificationSample(
            name=name, text=data["text"], label=data["label"], unique_id=unique_id
        )


class TokenClassificationSample(DataSamples):
    datatype = "token_classification"

    def __init__(
        self,
        name: str,
        tokens: typing.List[str],
        tags: typing.List[str],
        unique_id: str = None,
    ):
        super().__init__(unique_id=unique_id)

        self._tokens = tokens
        self._tags = tags
        self._name = name

    def serialize(self) -> str:
        return json.dumps({"tokens": self._tokens, "tags": self._tags})

    @staticmethod
    def deserialize(name, repr, unique_id):
        data = json.loads(repr)
        return TokenClassificationSample(
            tokens=data["tokens"], tags=data["tags"], name=name, unique_id=unique_id
        )


class XMLTokenClassificationSample(DataSamples):
    datatype = "xml_tokenclassification"

    @staticmethod
    def clean_rawlog(rawlog: str):
        start = rawlog.find("<")
        end = rawlog.rfind(">")
        if start == -1 or end == -1:
            raise Exception("Invalid XML Log")

        rawlog = rawlog[start : end + 1]

        clean_string = re.sub(r"[^\x20-\x7E\t\n\r]", "", rawlog)
        return clean_string

    @staticmethod
    def clean_fields(fields):
        for key, value in fields.items():
            if isinstance(value, dict):
                fields[key] = value["value"]

            fields[key] = str(fields[key])

        return fields

    def __init__(self, name: str, rawlog: str, fields: str, unique_id: str = None):
        super().__init__(unique_id=unique_id)

        self._name = name
        self.rawlog = XMLTokenClassificationSample.clean_rawlog(rawlog)
        self.fields = XMLTokenClassificationSample.clean_fields(fields)

    def serialize(self) -> str:
        return json.dumps({"rawlog": self.rawlog, "fields": json.dumps(self.fields)})

    @staticmethod
    def deserialize(name, repr, unique_id):
        data = json.loads(repr)

        return XMLTokenClassificationSample(
            name=name,
            rawlog=data["rawlog"],
            fields=json.loads(data["fields"]),
            unique_id=unique_id,
        )


class UserFeedBack(DataType):
    def __init__(self, sample_uuid):
        self._sample_uuid = sample_uuid

    @property
    def sample_uuid(self):
        return self._sample_uuid

    @abstractstaticmethod
    def deserialize(sample_uuid: str, name: str, repr: str):
        # can decode the data
        pass


class XMLTokenClassificationFeedBack(UserFeedBack):
    datatype = "xml_tokenclassification"

    def __init__(
        self,
        sample_uuid: str,
        name: str,
        xpath: str,
        attribute: str,
        delimiter: str,
        index_to_label: typing.Dict[int, str],
    ):
        super().__init__(sample_uuid)

        self._name = name
        self._xpath = xpath
        self._attribute = attribute
        self._delimiter = delimiter
        self._index_to_label = index_to_label

    def serialize(self) -> str:
        return json.dumps(
            {
                "xpath": self._xpath,
                "attribute": self._attribute,
                "delimiter": self._delimiter,
                "index_to_label": self._index_to_label,
            }
        )

    @staticmethod
    def deserialize(sample_uuid, name, repr) -> str:
        data = json.loads(repr)
        return XMLTokenClassificationFeedBack(
            name=name,
            xpath=data["xpath"],
            attribute=data["attribute"],
            delimiter=data["delimiter"],
            index_to_label=data["index_to_label"],
            sample_uuid=sample_uuid,
        )


def deserialize_sample_datatype(
    type: str, unique_id: str, name: str, serialized_data: str
):
    if type == "text_classification":
        return TextClassificationSample.deserialize(
            name, serialized_data, unique_id=unique_id
        )

    if type == "token_classification":
        return TokenClassificationSample.deserialize(
            name, serialized_data, unique_id=unique_id
        )

    if type == "xml_tokenclassification":
        return XMLTokenClassificationSample.deserialize(
            name, serialized_data, unique_id=unique_id
        )

    raise Exception(f"Cannot deserialize unknown datatype {type}")


def deserialize_userfeedback(
    type: str, sample_uuid: str, name: str, serialized_data: str
):
    if type == "xml_tokenclassification":
        return XMLTokenClassificationFeedBack.deserialize(
            sample_uuid=sample_uuid, name=name, repr=serialized_data
        )
