from pydantic import BaseModel
from typing import List, Dict, Optional


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
