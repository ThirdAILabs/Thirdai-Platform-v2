from enum import Enum
from typing import Optional

from pydantic import BaseModel, model_validator


class LLMProvider(str, Enum):
    openai = "openai"
    cohere = "cohere"
    onprem = "onprem"
    mock = "mock"


class LLMConfig(BaseModel):
    provider: LLMProvider
    api_key: Optional[str] = None
    base_url: Optional[str] = None
    model_name: Optional[str] = None

    @model_validator(mode="before")
    def check_llm_requirements(cls, values):
        provider = values.get("provider")
        api_key = values.get("api_key")
        model_name = values.get("model_name")

        if provider not in [LLMProvider.onprem, LLMProvider.mock]:
            if not api_key:
                raise ValueError("api_key must be provided for non-onprem providers")
        return values
