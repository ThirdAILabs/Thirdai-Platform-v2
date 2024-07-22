from abc import ABC, abstractmethod
from typing import Optional

from openai import OpenAI


class GenerativeBaseModel(ABC):
    def __init__(self):
        pass

    @abstractmethod
    def generate_output(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
    ) -> str:
        pass


class OpenAIClient:
    def __init__(self, api_key: str):
        super().__init__()
        self.client = OpenAI(api_key=api_key)

    def generate_output(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        model_name: str = "gpt-4o",
    ) -> str:
        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})

        messages.append({"role": "user", "content": prompt})

        response = self.client.chat.completions.create(
            model=model_name,
            messages=messages,
            temperature=0.8,
        )

        return response.choices[0].message.content
