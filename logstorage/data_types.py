from __future__ import annotations

import json
import typing
from abc import abstractmethod, abstractproperty, abstractstaticmethod
from dataclasses import dataclass

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


class XMLTokenClassificationSample(DataType):
    datatype = "xml_tokenclassification_sample"

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

    def __init__(self, name: str, rawlog: str, fields: str):
        self._name = name
        self.rawlog = XMLTokenClassificationSample.clean_rawlog(rawlog)
        self.fields = XMLTokenClassificationSample.clean_fields(fields)

    def serialize(self) -> str:
        return json.dumps({"rawlog": self.rawlog, "fields": json.dumps(self.fields)})

    @staticmethod
    def deserialize(name, repr):
        data = json.loads(repr)

        return XMLTokenClassificationSample(
            name=name, rawlog=data["rawlog"], fields=data["fields"]
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
