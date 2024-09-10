import os
from abc import ABC, abstractmethod
from collections import defaultdict
from pathlib import Path
from typing import Optional

import cohere
from openai import OpenAI


class LLMBase(ABC):
    def __init__(self, response_file: Optional[Path] = None):
        self.response_file = response_file
        self.usage = dict()

    @abstractmethod
    def completion(
        self, prompt: str, system_prompt: Optional[str] = None, **kwargs
    ) -> str:
        pass


class OpenAILLM(LLMBase):
    def __init__(self, api_key: str, response_file: Optional[Path] = None):
        super().__init__(response_file)
        self.client = OpenAI(api_key=api_key)

    def completion(
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
            temperature=0.8,  # TODO (anyone): Choose the temp based on a random distribution
        )

        res = response.choices[0].message.content
        usage = dict(response.usage)
        if self.response_file:
            with open(self.response_file, "a") as fp:
                fp.write(f"Prompt: \n{prompt}\n")
                fp.write(f"Response: \n{res}\n")
                fp.write(f"\nUsage: \n{usage}\n")
                fp.write("=" * 100 + "\n\n")

        self.usage = {
            key: self.usage.get(key, 0) + value for key, value in usage.items()
        }
        return res


class CohereLLM(LLMBase):
    def __init__(self, api_key: str, response_file: Optional[Path] = None):
        super().__init__(response_file)
        self.client = cohere.Client(api_key=api_key)

    def completion(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        model_name: str = "command-r-plus",
    ) -> str:
        message = ""
        if system_prompt:
            message = f"{system_prompt}\n\n"
        message += prompt

        response = self.client.chat(model=model_name, message=message)

        if self.response_file:
            with open(self.response_file, "a") as fp:
                fp.write(f"Prompt: \n{prompt}\n")
                fp.write(f"Response: \n{response.text}\n")
                fp.write("=" * 100 + "\n\n")

        return response.text


llm_classes = {
    "openai": OpenAILLM,
    "cohere": CohereLLM,
}
