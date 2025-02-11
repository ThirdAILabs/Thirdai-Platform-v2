from enum import Enum
from typing import Optional

from pydantic import BaseModel


class LLMProvider(str, Enum):
    openai = "openai"
    cohere = "cohere"
    onprem = "on-prem"


class LLMConfig(BaseModel):
    provider: LLMProvider
    api_key: str
    base_url: Optional[str] = None
    model_name: str
