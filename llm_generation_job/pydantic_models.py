from typing import Optional

from pydantic import BaseModel


class GenerateArgs(BaseModel):
    query: str
    system_prompt: Optional[str] = None
    key: Optional[str] = None
    model: str = "gpt-3.5-turbo"
    provider: str = "openai"
