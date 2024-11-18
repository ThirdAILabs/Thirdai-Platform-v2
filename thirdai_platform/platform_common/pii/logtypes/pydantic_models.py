from pydantic import BaseModel
from typing import List, Dict, Literal, Optional


class XPathLocation(BaseModel):
    xpath: str
    attribute: Optional[str]


class CharSpan(BaseModel):
    start: int
    end: int


class XMLLocation(BaseModel):
    char_span: CharSpan
    xpath_location: XPathLocation
    value: str


class XMLPrediction(BaseModel):
    label: str
    location: XMLLocation


class XMLTokenClassificationResults(BaseModel):
    literal: Literal["xml"]
    query_text: str
    predictions: List[XMLPrediction]


class UnstructuredTokenClassificationResults(BaseModel):
    literal: Literal["unstructured"]
    query_text: str
    tokens: List[str]
    predicted_tags: List[List[str]]
