from typing import Dict, List, Optional

from pydantic import BaseModel


class XpathLocation(BaseModel):
    xpath: str
    attr: Optional[str]


class CharSpan(BaseModel):
    start: int
    end: int


class Prediction(BaseModel):
    xpath_location: XpathLocation
    char_span: CharSpan
    label: str
    value: str


class XMLPredictions(BaseModel):
    predictions: List[Prediction]
