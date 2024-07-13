from abc import ABC, abstractmethod
from pathlib import Path

from reporter import Reporter
from variables import GeneralVariables, TrainVariables


class Model(ABC):
    def __init__(self):
        self.general_variables = GeneralVariables.load_from_env()
        self.train_variables = TrainVariables.load_from_env()
        self.reporter = Reporter(self.general_variables.model_bazaar_endpoint)
        self.data_dir = (
            Path(self.general_variables.model_bazaar_dir)
            / "data"
            / self.general_variables.data_id
        )
        self.data_dir.mkdir(parents=True, exist_ok=True)

        self.model_dir = (
            Path(self.general_variables.model_bazaar_dir)
            / "models"
            / self.general_variables.model_id
        )
        self.model_dir.mkdir(parents=True, exist_ok=True)

    @abstractmethod
    def train(self, **kwargs):
        pass

    @abstractmethod
    def evaluate(self, **kwargs):
        pass
