"""
Defines the abstract base class for models.
"""

import json
from abc import ABC, abstractmethod
from pathlib import Path

from variables import GeneralVariables


class Model(ABC):
    """
    Abstract base class for all models.
    """

    def __init__(self) -> None:
        """
        Initializes model directories and reporter.
        """
        self.general_variables: GeneralVariables = GeneralVariables.load_from_env()
        self.model_dir: Path = self.get_model_dir()

    @abstractmethod
    def predict(self, **kwargs):
        """
        Abstract method for prediction.
        """
        pass

    def get_model_dir(self) -> Path:
        """
        Returns the directory path for the given model ID.
        """
        return Path(self.general_variables.checkpoint_dir)
