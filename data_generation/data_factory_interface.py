import os
from abc import ABC, abstractmethod

from api_clients import OpenAIClient
from variables import GeneralVariables


class DataFactory(ABC):
    def __init__(self, api_key: str):
        self.general_variables: GeneralVariables = GeneralVariables.load_from_env()
        self.save_dir = os.path.join(
            self.general_variables.model_bazaar_dir, self.general_variables.data_id
        )
        os.makedirs(self.save_dir, exist_ok=True)

        self.openai = OpenAIClient(api_key=api_key)

    @abstractmethod
    def generate(self, **kwargs):
        pass
