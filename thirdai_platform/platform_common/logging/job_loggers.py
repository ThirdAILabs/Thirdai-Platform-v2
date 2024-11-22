import logging
from pathlib import Path
from typing import Dict, List, Union

from platform_common.logging.base_logger import BaseLogger


class DeploymentLogger(BaseLogger):
    def __init__(
        self,
        log_dir: Path,
        log_prefix: str,
        model_id: str,
        model_type: str,
        user_id: str,
        service_type: str = "deployment",
        level: int = logging.INFO,
    ):
        super().__init__(log_dir, log_prefix, service_type, level)
        self.model_id = model_id
        self.user_id = user_id
        self.model_type = model_type

    @property
    def get_logger_keys(self) -> Dict[str, Union[str, int]]:
        return {
            "model_id": self.model_id,
            "user_id": self.user_id,
            "model_type": self.model_type,
        }

    @property
    def stream_fields(self) -> List[str]:
        # skipping user_id because a large number of users can access the same deployment
        return ["model_id", "model_type"]


class TrainingLogger(BaseLogger):
    def __init__(
        self,
        log_dir: Path,
        log_prefix: str,
        model_id: str,
        model_type: str,
        user_id: str,
        service_type: str = "training",
        level: int = logging.INFO,
    ):
        super().__init__(log_dir, log_prefix, service_type, level)
        self.model_id = model_id
        self.user_id = user_id
        self.model_type = model_type

    @property
    def get_logger_keys(self) -> Dict[str, Union[str, int]]:
        return {
            "model_id": self.model_id,
            "user_id": self.user_id,
            "model_type": self.model_type,
        }

    @property
    def stream_fields(self) -> List[str]:
        # skipping user_id because a large number of users can access the same deployment
        return ["model_id", "model_type"]
