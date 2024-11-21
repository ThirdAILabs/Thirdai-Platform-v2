import json
import logging
from datetime import datetime
from pathlib import Path

from colorlog import ColoredFormatter


class JSONFormatter(logging.Formatter):
    """Formatter that outputs JSON strings with datetime for log files"""

    def format(self, record) -> str:
        # Format the time as string
        timestamp = datetime.fromtimestamp(record.created).strftime("%Y-%m-%d %H:%M:%S")

        # Basic log entry with required fields
        log_entry = {
            "timestamp": timestamp,
            "level": record.levelname,
            "logger_name": record.name,
            "message": str(record.msg),
        }

        # Add code if present
        if hasattr(record, "log_code"):
            log_entry["code"] = record.log_code

        # Add extra fields if present
        if hasattr(record, "extra_fields"):
            log_entry.update(record.extra_fields)

        return json.dumps(log_entry)


class BaseLogger:
    """Abstract base class for all loggers"""

    def __init__(
        self,
        log_dir: Path,
        log_prefix: str,
        service_name: str = "default",
        level=logging.INFO,
    ):
        self.logger = self._setup_logger(log_dir, log_prefix, service_name, level)

    def _setup_logger(
        self, log_dir: Path, log_prefix: str, service_name: str, level: int
    ) -> logging.Logger:
        log_dir.mkdir(parents=True, exist_ok=True)
        logger_file_path = log_dir / f"{log_prefix}.log"

        logger = logging.getLogger(service_name)
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
