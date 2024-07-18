import os
from logger.log_module import LoggerConfig

log_filepath = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "logfile")
logger = LoggerConfig(os.getenv("MODEL_BAZAAR_DIR")).get_logger(log_filepath)
