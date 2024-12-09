from platform_common.pii.data_types.base import LogType
from platform_common.pii.data_types.pydantic_models import (
    UnstructuredTokenClassificationResults,
    XMLTokenClassificationResults,
)
from platform_common.pii.data_types.unstructured import UnstructuredText
from platform_common.pii.data_types.xml import XMLTokenClassificationLog


def convert_log_to_concrete_type(log: str):
    try:
        return XMLTokenClassificationLog(log)
    except:
        return UnstructuredText(log)


__all__ = [
    "DataType",
    "UnstructuredText",
    "XMLTokenClassificationLog",
    "XMLTokenClassificationResults",
    "UnstructuredTokenClassificationResults",
]
