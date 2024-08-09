import csv
import json
import random
import traceback
from abc import ABC, abstractmethod
from pathlib import Path
from resource.util_data import random_prompts, vocab
from typing import Dict, List, Optional

from llms import llm_classes
from variables import GeneralVariables


class DataFactory(ABC):
    def __init__(self):
        self.general_variables: GeneralVariables = GeneralVariables.load_from_env()
        self.save_dir = (
            Path(self.general_variables.model_bazaar_dir)
            / self.general_variables.data_id
        )
        self.save_dir.mkdir(parents=True, exist_ok=True)

        self.train_file_location = self.save_dir / "train.csv"
        self.errored_file_location = self.save_dir / "traceback.err"
        self.config_file_location = self.save_dir / "config.json"
        self.generation_args_location = self.save_dir / "generation_args.json"

    def init_llm(self):
        return llm_classes.get(self.general_variables.llm_provider.value)(
            api_key=self.general_variables.genai_key
        )

    def llm_completion(self, prompt: str, system_prompt: Optional[str] = None):
        llm_model = self.init_llm()
        return llm_model.completion(prompt, system_prompt=system_prompt)

    @abstractmethod
    def generate_data(self, **kwargs):
        pass

    @abstractmethod
    def fill_and_transform(self, **kwargs):
        pass

    def get_random_vocab(self, user_vocab: Optional[str] = None, k=1):
        vocabulary = vocab + (user_vocab if user_vocab is not None else [])
        return random.sample(population=vocabulary, k=k)

    def get_random_prompts(self, k=1):
        return [
            random.choices(items["prompts"], weights=items["scores"], k=k)[0]
            for __annotations__, items in random_prompts.items()
        ]

    def write_on_training_file(
        self,
        data_points: List[Dict[str, str]],
        fieldnames: List[str],
        write_fields: bool = True,
        newline: Optional[str] = None,
        encoding: Optional[str] = None,
    ):
        try:
            with open(
                self.train_file_location, "a", newline=newline, encoding=encoding
            ) as csv_file:
                csv_writer = csv.DictWriter(csv_file, fieldnames=fieldnames)
                if write_fields:
                    csv_writer.writeheader()
                csv_writer.writerows(data_points)
        except Exception as e:
            with open(self.errored_file_location, mode="a") as errored_fp:
                errored_fp.write(
                    "\nError while writing on train file " + "-" * 20 + "\n"
                )
                traceback.print_exc(file=errored_fp)
                errored_fp.write("\n" + "=" * 100 + "\n")

    def save_dict(self, write_to: str, **kwargs):
        with open(write_to, "w") as fp:
            json.dump(kwargs, fp, indent=4)
