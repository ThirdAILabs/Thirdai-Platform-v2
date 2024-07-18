import os

from logging_service import LoggerConfig
from variables import GeneralVariables

general_variables: GeneralVariables = GeneralVariables.load_from_env()

model_dir = os.path.join(
    general_variables.model_bazaar_dir, "models", general_variables.model_id
)
os.makedirs(model_dir, exist_ok=True)
logger = LoggerConfig(os.path.join(model_dir, "train.log")).get_logger("train-logger")
logger.info("Starts train logging")
