from __future__ import annotations

import json
import typing
from abc import abstractmethod, abstractproperty, abstractstaticmethod
from dataclasses import dataclass

import pandas as pd


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


class TextClassificationSample(DataType):
    datatype = "text_classification"

    def __init__(self, name: str, text: str, label: str):
        self._text = text
        self._label = label
        self._name = name

    def serialize(self) -> str:
        return json.dumps({"text": self._text, "label": self._label})

    @staticmethod
    def deserialize(name, repr) -> str:
        data = json.loads(repr)
        return TextClassificationSample(
            tokens=data["text"], label=data["label"], name=name
        )


class TokenClassificationSample(DataType):
    datatype = "token_classification"

    def __init__(self, name: str, tokens: typing.List[str], tags: typing.List[str]):
        self._tokens = tokens
        self._tags = tags
        self._name = name

    def serialize(self) -> str:
        return json.dumps({"tokens": self._tokens, "tags": self._tags})

    @staticmethod
    def deserialize(name, repr):
        data = json.loads(repr)
        return TokenClassificationSample(
            tokens=data["tokens"], tags=data["tags"], name=name
        )


class XMLTokenClassificationFeedBack(DataType):
    datatype = "xml_tokenclassification_userfeedback"

    def __init__(
        self,
        name: str,
        xpath: str,
        attribute: str,
        delimiter: str,
        index_to_label: typing.Dict[int, str],
    ):
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
    def deserialize(name, repr) -> str:
        data = json.loads(repr)
        return XMLTokenClassificationFeedBack(
            name=name,
            xpath=data["xpath"],
            attribute=data["attribute"],
            delimiter=data["delimiter"],
            index_to_label=data["index_to_label"],
        )


def deserialize_datatype(type: str, name: str, serialized_data: str):
    if type == "text_classification":
        return TextClassificationSample.deserialize(name, serialized_data)

    if type == "token_classification":
        return TokenClassificationSample.deserialize(name, serialized_data)

    if type == "xml_tokenclassification_userfeedback":
        return XMLTokenClassificationFeedBack.deserialize(name, serialized_data)

    raise Exception(f"Cannot deserialize unknown datatype {type}")
