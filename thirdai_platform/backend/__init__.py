import os

from logger.service import LoggerConfig

log_filepath = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "logfile")
logger = LoggerConfig(log_filepath).get_logger("bazaar-logger")
