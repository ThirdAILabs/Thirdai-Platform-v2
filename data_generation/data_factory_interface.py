import json
import random
from abc import ABC, abstractmethod
from pathlib import Path
from resource.util_data import random_prompts, vocab
from typing import Optional

from variables import GeneralVariables

from .llms import llm_classes


class DataFactory(ABC):
    def __init__(self):
        self.general_variables: GeneralVariables = GeneralVariables.load_from_env()
        self.save_dir = (
            Path(self.general_variables.model_bazaar_dir)
            / self.general_variables.data_id
        )
        self.save_dir.mkdir(parents=True, exist_ok=True)

        self.llm_model = llm_classes.get(self.general_variables.llm_provider.value)(
            api_key=self.general_variables.genai_key
        )

    @abstractmethod
    def generate_data(self, **kwargs):
        pass

    def get_random_vocab(self, user_vocab: Optional[str] = None, k=1):
        vocabulary = vocab + (user_vocab if user_vocab is not None else [])
        return random.sample(population=vocabulary, k=k)

    def get_random_prompts(self, k=1):
        return [
            random.choices(items["prompts"], weights=items["scores"], k=k)[0]
            for __annotations__, items in random_prompts.items()
        ]

    def save_config(self, **kwargs):
        with open(self.save_dir / "config.json", "w") as config_fp:
            json.dump(kwargs, config_fp, indent=4)
