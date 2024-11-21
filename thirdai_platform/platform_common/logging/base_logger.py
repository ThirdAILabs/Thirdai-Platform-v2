import json
import logging
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Union

from colorlog import ColoredFormatter
from platform_common.logging.logcodes import LogCode

"""
Similar to Loki, VictoriaLogs supports full-text search over the logs. Since, the number of logs can be very large,
it also supports assigning (key,value) pairs to the logs which can be used for filtering the logs.

To make search efficient and more customizable, it also supports marking certain fields as "stream fields".
Searching over stream fields is significantly faster than doing a full text search or simple key-value pair search.

BaseLogger and all its derived classes should have two methods:
1. get_logger_keys: Returns a dictionary of key,value pairs enabling searching over logs.
2. stream_fields: Returns a list of fields which should be marked as stream fields in VictoriaLogs.

The fields in stream_fields should not have very high cardinality (i.e. number of unique values should be limited).

Check out this doc to read about how VictoriaLogs stores and retrieves logs:
https://docs.victoriametrics.com/victorialogs/keyconcepts/#stream-fields
"""


class JSONFormatter(logging.Formatter):
    """Formatter that outputs JSON strings with datetime for log files"""

    def format(self, record) -> str:
        # Format the time as string
        timestamp = datetime.fromtimestamp(record.created).strftime("%Y-%m-%d %H:%M:%S")

        # Basic log entry with required fields
        log_entry = {
            "level": record.levelname,
            "service_type": record.name,
            # reserved fields name for VictoriaLogs
            "_time": timestamp,
            "_msg": str(record.msg),
        }

        # Add code if present
        if hasattr(record, "code"):
            log_entry["code"] = record.code

        # Add extra fields if present
        if hasattr(record, "extra_fields"):
            log_entry.update(record.extra_fields)

        return json.dumps(log_entry)


class BaseLogger:
    """Base class for all loggers"""

    def __init__(
        self,
        log_dir: Path,
        log_prefix: str,
        service_type: str,
        level=logging.INFO,
    ):
        self.logger = self._setup_logger(log_dir, log_prefix, service_type, level)

    @property
    def get_logger_keys(self) -> Dict[str, Union[str, int]]:
        """Returns key,value pairs that can be used to filter logs"""
        return {}

    @property
    def stream_fields(self) -> List[str]:
        """Returns keys used for indexing logs into unique streams in VictoriaLogs"""
        return []

    def _setup_logger(
        self, log_dir: Path, log_prefix: str, service_type: str, level: int
    ) -> logging.Logger:
        log_dir.mkdir(parents=True, exist_ok=True)
        logger_file_path = log_dir / f"{log_prefix}.log"

        logger = logging.getLogger(service_type)
        logger.setLevel(level)
        logger.handlers = []

        file_handler = logging.FileHandler(logger_file_path, mode="a+")
        file_handler.setFormatter(JSONFormatter())

        console_formatter = ColoredFormatter(
            "%(log_color)s%(asctime)s - %(levelname)s - %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
            log_colors={
                "DEBUG": "cyan",
                "INFO": "green",
                "WARNING": "yellow",
                "ERROR": "red",
                "CRITICAL": "red,bg_white",
            },
        )
        console_handler = logging.StreamHandler()
        console_handler.setFormatter(console_formatter)

        logger.addHandler(file_handler)
        logger.addHandler(console_handler)
        return logger

    def _log(self, level: int, code: str, message: str, **extra_fields):
        """Basic logging method that only takes code, message and extra fields"""
        extra = {"code": code, "extra_fields": extra_fields}
        self.logger.log(level, message, extra=extra)

    def info(self, code: LogCode, message: str, **extra_fields):
        """Log an info message with code and extra fields"""
        logger_keys = self.get_logger_keys
        extra_fields.update(logger_keys)
        self._log(logging.INFO, code, message, **extra_fields)

    def debug(self, code: LogCode, message: str, **extra_fields):
        """Log a debug message with code and extra fields"""
        logger_keys = self.get_logger_keys
        extra_fields.update(logger_keys)
        self._log(logging.DEBUG, code, message, **extra_fields)

    def warning(self, code: LogCode, message: str, **extra_fields):
        """Log a warning message with code and extra fields"""
        logger_keys = self.get_logger_keys
        extra_fields.update(logger_keys)
        self._log(logging.WARNING, code, message, **extra_fields)

    def error(self, code: LogCode, message: str, **extra_fields):
        """Log an error message with code and extra fields"""
        logger_keys = self.get_logger_keys
        extra_fields.update(logger_keys)
        self._log(logging.ERROR, code, message, **extra_fields)

    def critical(self, code: LogCode, message: str, **extra_fields):
        """Log a critical message with code and extra fields"""
        logger_keys = self.get_logger_keys
        extra_fields.update(logger_keys)
        self._log(logging.CRITICAL, code, message, **extra_fields)
