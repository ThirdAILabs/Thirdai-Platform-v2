from abc import ABC, abstractmethod
from pathlib import Path

from reporter import Reporter
from variables import GeneralVariables


class Model(ABC):
    def __init__(self):
        self.general_variables = GeneralVariables.load_from_env()
        self.reporter = Reporter(self.general_variables.model_bazaar_endpoint)

        self.model_dir = self.get_model_dir(model_id=self.general_variables.model_id)

        self.data_dir = (
            self.model_dir
            / "deployments"
            / self.general_variables.deployment_id
            / "data"
        )
        self.data_dir.mkdir(parents=True, exist_ok=True)

    @abstractmethod
    def predict(self, **kwargs):
        pass

    def get_model_dir(self, model_id) -> Path:
        return Path(self.general_variables.model_bazaar_dir) / "models" / model_id
