import os
import re
import warnings
from typing import Any, Dict, List, Optional

from thirdai import neural_db as ndb

from headless.configs import Config


def get_csv_source_id(
    file: str,
    CSV_ID_COLUMN: Optional[str] = None,
    CSV_STRONG_COLUMNS: Optional[List[str]] = None,
    CSV_WEAK_COLUMNS: Optional[List[str]] = None,
    CSV_REFERENCE_COLUMNS: Optional[List[str]] = None,
    CSV_METADATA: Optional[Dict[str, str]] = None,
) -> str:
    """
    Returns the source ID for a CSV file.

    Parameters:
    file (str): Path to the CSV file.
    CSV_ID_COLUMN (str, optional): Column name for IDs.
    CSV_STRONG_COLUMNS (list[str], optional): List of strong columns.
    CSV_WEAK_COLUMNS (list[str], optional): List of weak columns.
    CSV_REFERENCE_COLUMNS (list[str], optional): List of reference columns.
    CSV_METADATA (dict[str, str], optional): Metadata for the CSV file.

    Returns:
    str: The hash ID of the CSV source.

    Raises:
    TypeError: If the file type is not supported.
    """
    _, ext = os.path.splitext(file)

    if ext == ".csv":
        return ndb.CSV(
            file,
            id_column=CSV_ID_COLUMN,
            strong_columns=CSV_STRONG_COLUMNS,
            weak_columns=CSV_WEAK_COLUMNS,
            reference_columns=CSV_REFERENCE_COLUMNS,
            metadata=CSV_METADATA,
        ).hash
    else:
        raise TypeError(f"{ext} Document type isn't supported.")


def build_extra_options(config: Config, sharded: bool = False) -> Dict[str, Any]:
    """
    Builds a dictionary of extra options for training.

    Parameters:
    config (Config): Configuration object containing various settings.
    sharded (bool, optional): Whether to use sharded training.

    Returns:
    dict[str, Any]: Dictionary of extra training options.
    """
    return {
        "model_cores": config.model_cores,
        "model_memory": config.model_memory,
        "csv_id_column": config.id_column,
        "csv_strong_columns": config.strong_columns,
        "csv_weak_columns": config.weak_columns,
        "csv_reference_columns": config.reference_columns,
        "fhr": config.input_dim,
        "embedding_dim": config.hidden_dim,
        "output_dim": config.output_dim,
        "csv_query_column": config.query_column,
        "csv_id_delimiter": config.id_delimiter,
        "num_models_per_shard": 2 if sharded else 1,
        "num_shards": 2 if sharded else 1,
        "allocation_memory": config.allocation_memory,
        "unsupervised_epochs": config.epochs,
        "supervised_epochs": config.epochs,
        "retriever": config.retriever,
    }


def get_configs(config_type: type, config_regex: str) -> List[Config]:
    """
    Retrieves a list of configuration subclasses that match a given regex pattern.

    Parameters:
    config_type (type): The base configuration class type.
    config_regex (str): Regular expression to filter configuration names.

    Returns:
    list[Config]: List of matching configuration subclasses.

    Raises:
    Warning: If no configurations match the regex pattern.
    """
    configs = [config for config in config_type.__subclasses__()]
    config_re = re.compile(config_regex)
    configs = list(
        filter(
            lambda config: config.name is not None and config_re.match(config.name),
            configs,
        )
    )
    if len(configs) == 0:
        warnings.warn(
            f"Couldn't match regular expression '{config_regex}' to any configs"
        )

    return configs


def create_doc_dict(path: str, doc_type: str) -> Dict[str, str]:
    """
    Creates a document dictionary for different document types.

    Parameters:
    path (str): Path to the document file.
    doc_type (str): Type of the document location.

    Returns:
    dict[str, str]: Dictionary containing document details.

    Raises:
    Exception: If the document type is not supported.
    """
    _, ext = os.path.splitext(path)
    if ext == ".pdf":
        return {"document_type": "PDF", "path": path, "location": doc_type}
    if ext == ".csv":
        return {"document_type": "CSV", "path": path, "location": doc_type}
    if ext == ".docx":
        return {"document_type": "DOCX", "path": path, "location": doc_type}

    raise Exception(f"Please add a map from {ext} to document dictionary.")
