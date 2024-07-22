import os
import logging

from stream_logger.service import LoggerConfig

log_filepath = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "logfile")
logging_levels = {
    'debug': logging.debug,
    'info': logging.info,
    'warning': logging.warning,
    'error': logging.error,
    'critical': logging.critical
}
log_level = os.getenv('LOG_LEVEL', 'info')
assert log_level in logging_levels
logger = LoggerConfig(log_filepath, logging_levels[log_level]).get_logger("bazaar-logger")
