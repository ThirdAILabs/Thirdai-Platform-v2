import json
from typing import Dict, List, Optional

from pydantic import BaseModel, Field
from pydantic_models.constraints import Constraints


class AssociateInputSingle(BaseModel):
    source: str
    target: str


class AssociateInput(BaseModel):
    text_pairs: List[AssociateInputSingle]


class Reference(BaseModel):
    id: int = Field(..., ge=0)
    text: str
    context: str
    source: str
    metadata: Dict
    source_id: str
    score: float


class UpvoteInputSingle(BaseModel):
    query_text: str
    reference_id: int = Field(..., ge=0)


class UpvoteInput(BaseModel):
    text_id_pairs: List[UpvoteInputSingle]


class SearchResults(BaseModel):
    query_text: str
    references: List[Reference]


class DeleteInput(BaseModel):
    source_ids: List[str]


class SaveModel(BaseModel):
    override: bool
    model_name: Optional[str] = None


def convert_reference_to_pydantic(input, context_radius: int):
    def convert_to_json_encodable(value):
        try:
            json_value = json.dumps(value)
            return json_value
        except TypeError:
            return str(value)

    return Reference(
        id=input.id,
        text=input.text,
        source=input.source,
        metadata={k: convert_to_json_encodable(v) for k, v in input.metadata.items()},
        context=input.context(radius=context_radius),
        source_id=input.document.hash,
        score=input.score,
    )


class BaseQueryParams(BaseModel):
    query: str
    top_k: int = 5


class NDBExtraParams(BaseModel):
    rerank: bool = False
    top_k_rerank: int = 100
    context_radius: int = 1
    rerank_threshold: float = 1.5
    top_k_threshold: Optional[int] = None
    constraints: Constraints = Field(default_factory=Constraints)
