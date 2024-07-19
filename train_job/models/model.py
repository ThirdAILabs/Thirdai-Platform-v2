from abc import ABC, abstractmethod
from pathlib import Path

from reporter import Reporter
from variables import GeneralVariables, TrainVariables


class Model(ABC):
    """
    Abstract base class for a model.
    Provides common initialization and abstract methods for training and evaluation.
    """

    def __init__(self):
        """
        Initialize the model with general and training variables,
        create necessary directories, and set up a reporter for status updates.
        """
        self.general_variables: GeneralVariables = GeneralVariables.load_from_env()
        self.train_variables: TrainVariables = TrainVariables.load_from_env()
        self.reporter: Reporter = Reporter(self.general_variables.model_bazaar_endpoint)

        # Directory for storing data
        self.data_dir: Path = (
            Path(self.general_variables.model_bazaar_dir)
            / "data"
            / self.general_variables.data_id
        )
        self.data_dir.mkdir(parents=True, exist_ok=True)

        # Directory for storing model outputs
        self.model_dir: Path = (
            Path(self.general_variables.model_bazaar_dir)
            / "models"
            / self.general_variables.model_id
        )
        self.model_dir.mkdir(parents=True, exist_ok=True)

        self.unsupervised_checkpoint_dir: Path = (
            self.model_dir / "checkpoints" / "unsupervised"
        )
        self.supervised_checkpoint_dir: Path = (
            self.model_dir / "checkpoints" / "supervised"
        )

    @abstractmethod
    def train(self, **kwargs):
        """
        Abstract method for training the model. Must be implemented by subclasses.
        """
        pass

    @abstractmethod
    def evaluate(self, **kwargs):
        """
        Abstract method for evaluating the model. Must be implemented by subclasses.
        """
        pass
