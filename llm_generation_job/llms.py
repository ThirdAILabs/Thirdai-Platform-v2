import json
import os
from typing import AsyncGenerator

import aiohttp


class LLMBase:
    async def stream(
        self, key: str, query: str, model: str
    ) -> AsyncGenerator[str, None]:
        raise NotImplementedError("Subclasses must implement this method")


class OpenAILLM(LLMBase):
    async def stream(
        self, key: str, query: str, model: str
    ) -> AsyncGenerator[str, None]:
        url = "https://api.openai.com/v1/chat/completions"
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {key}",
        }
        body = {
            "model": model,
            "messages": [
                {
                    "role": "user",
                    "content": query,
                }
            ],
            "stream": True,
        }
        async with aiohttp.ClientSession() as session:
            async with session.post(url, headers=headers, json=body) as response:
                if response.status == 200:
                    async for multi_chunk_bytes, _ in response.content.iter_chunks():
                        for chunk_string in multi_chunk_bytes.decode("utf8").split(
                            "\n"
                        ):
                            if chunk_string == "":
                                continue
                            offset = len(
                                "data: "
                            )  # The chunk responses are prefixed with "data: "
                            try:
                                chunk = json.loads(chunk_string[offset:])
                            except:
                                continue
                            content = (
                                chunk["choices"][0].get("delta", {}).get("content")
                            )
                            if content is not None:
                                yield content


class CohereLLM(LLMBase):
    async def stream(
        self, key: str, query: str, model: str
    ) -> AsyncGenerator[str, None]:
        url = "https://api.cohere.com/v1/chat"
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {key}",
        }
        body = {
            "message": query,
            "model": model,
            "chat_history": [
                {
                    "role": "USER",
                    "message": query,
                }
            ],
            "stream": True,
        }
        async with aiohttp.ClientSession() as session:
            async with session.post(url, headers=headers, json=body) as response:
                if response.status == 200:
                    async for line in response.content:
                        line = line.decode("utf8").strip()
                        try:
                            chunk = json.loads(line)
                            if chunk.get(
                                "event_type"
                            ) == "text-generation" and not chunk.get("is_finished"):
                                content = chunk.get("text")
                                if content:
                                    yield content
                        except json.JSONDecodeError as e:
                            raise Exception(f"Error decoding JSON response: {e}")
                        except Exception as e:
                            raise Exception(f"Error processing response chunk: {e}")
                else:
                    error_message = await response.text()
                    raise Exception(f"Cohere API request failed: {error_message}")


model_classes = {
    "openai": OpenAILLM,
    "cohere": CohereLLM,
}

default_keys = {
    "openai": os.getenv("OPENAI_KEY", ""),
    "cohere": os.getenv("COHERE_KEY", ""),
}