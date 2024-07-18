import os

from logging_service import LoggerConfig
from variables import GeneralVariables

general_variables: GeneralVariables = GeneralVariables.load_from_env()

log_file = os.path.join(
    general_variables.model_bazaar_dir,
    "models",
    general_variables.model_id,
    f"deployement-{general_variables.deployment_id}.log",
)
logger = LoggerConfig(log_file).get_logger("train-logger")
logger.info("Starts deployment logging")
