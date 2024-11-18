from platform_common.pii.logtypes.xml import XMLTokenClassificationLog
from platform_common.pii.logtypes.unstructured import UnstructuredTokenClassificationLog


def convert_log_to_concrete_type(log: str):
    try:
        return XMLTokenClassificationLog(log)
    except:
        return UnstructuredTokenClassificationLog(log)
