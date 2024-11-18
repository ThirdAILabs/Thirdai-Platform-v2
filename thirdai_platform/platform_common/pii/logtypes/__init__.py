from platform_common.pii.logtypes.base import LogType
from platform_common.pii.logtypes.pydantic_models import (
    UnstructuredTokenClassificationResults,
    XMLTokenClassificationResults,
)
from platform_common.pii.logtypes.unstructured import UnstructuredTokenClassificationLog
from platform_common.pii.logtypes.xml import XMLTokenClassificationLog


def convert_log_to_concrete_type(log: str):
    try:
        return XMLTokenClassificationLog(log)
    except:
        return UnstructuredTokenClassificationLog(log)


__all__ = [
    "LogType",
    "UnstructuredTokenClassificationLog",
    "XMLTokenClassificationLog",
    "XMLTokenClassificationResults",
    "UnstructuredTokenClassificationResults",
]
