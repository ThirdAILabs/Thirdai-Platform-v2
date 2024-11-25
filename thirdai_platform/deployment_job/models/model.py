from abc import ABC
from pathlib import Path

from platform_common.logging import DeploymentLogger
from platform_common.logging.logcodes import LogCode
from platform_common.pydantic_models.deployment import DeploymentConfig

# Logging Levels for Concrete Models are : Debug and Error


class Model(ABC):
    """
    Abstract base class for all models.
    """

    def __init__(self, config: DeploymentConfig, logger: DeploymentLogger) -> None:
        """
        Initializes model directories and reporter.
        """
        self.config = config
        self.logger = logger

        self.model_dir = self.get_model_dir(self.config.model_id)

        self.logger.debug(
            f"Model initialized with model_id: {self.config.model_id} at {self.model_dir}",
            code=LogCode.MODEL_INIT,
        )
        self.data_dir = self.model_dir / "deployments" / "data"

        self.data_dir.mkdir(parents=True, exist_ok=True)

        self.logger.debug(f"Data directory created or exists at {self.data_dir}")

    def get_model_dir(self, model_id: str):
        return Path(self.config.model_bazaar_dir) / "models" / model_id
