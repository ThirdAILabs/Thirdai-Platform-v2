import json
from abc import ABC, abstractmethod
from pathlib import Path
from typing import Optional

import websockets
from variables import GeneralVariables

from .utils import random_prompts


class DataFactory(ABC):
    def __init__(self, api_key: str):
        self.general_variables: GeneralVariables = GeneralVariables.load_from_env()
        self.save_dir = Path
        self.save_dir = (
            Path(self.general_variables.model_bazaar_dir)
            / self.general_variables.data_id
        )
        self.save_dir.mkdir(parents=True, exist_ok=True)

        self.api_key = api_key
        self.random_prompts = random_prompts

    @abstractmethod
    def generate_data(self, **kwargs):
        pass

    async def llm_completion(self, content: str, system_prompt: Optional[str] = None):
        uri = "ws://llm-generation-container:8000/generate"

        async with websockets.connect(uri) as websocket:
            args = {
                "query": content,
                "key": self.api_key,
                "model": "gpt-4o",
                "provider": self.general_variables.provider,
            }
            if system_prompt:
                args["system_prompt"] = system_prompt

            await websocket.send(json.dumps(args))

            response_text = ""
            while True:
                response = await websocket.receive_text()
                response_data = json.loads(response)

                response_text += response_data["content"]

                if response_data["end_of_stream"]:
                    return response_text
