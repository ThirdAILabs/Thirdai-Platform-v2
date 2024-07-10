from __future__ import annotations

import ast
import os
from dataclasses import MISSING, dataclass, fields
from enum import Enum
from typing import Dict, Optional, Type, TypeVar, Union, get_args, get_origin

T = TypeVar("T", bound="EnvLoader")


class EnvLoader:
    @classmethod
    def load_from_env(cls: Type[T]) -> T:
        missing_vars = []
        env_vars: Dict[str, Optional[Union[str, int, float, bool]]] = {}

        for f in fields(cls):
            value = os.getenv(f.name.upper())
            if value is None:
                if f.default is MISSING and f.default_factory is MISSING:
                    missing_vars.append(f.name.upper())
                else:
                    value = (
                        f.default if f.default is not MISSING else f.default_factory()
                    )
            else:
                value = cls._convert_type(value, f.type)
            env_vars[f.name] = value

        if missing_vars:
            raise ValueError(
                f"Missing required environment variables: {', '.join(missing_vars)}"
            )

        return cls(**env_vars)

    @staticmethod
    def _convert_type(
        value: str, field_type: Type
    ) -> Union[str, int, float, bool, None, Enum]:
        origin = get_origin(field_type)
        args = get_args(field_type)

        if origin is Union and type(None) in args:
            non_none_type = next(arg for arg in args if arg is not type(None))
            return (
                None
                if value.lower() == "none"
                else EnvLoader._convert_type(value, non_none_type)
            )

        if issubclass(field_type, Enum):
            return field_type(value)

        if field_type == bool:
            return ast.literal_eval(value.capitalize())
        if field_type == int:
            return int(value)
        if field_type == float:
            return float(value)
        if field_type == str:
            return value

        return ast.literal_eval(value)


class TypeEnum(str, Enum):
    NDB = "ndb"
    UDT = "udt"


class RetrieverEnum(str, Enum):
    MACH = "mach"
    HYBRID = "hybrid"
    FINETUNABLE_RETRIEVER = "finetunable_retriever"


@dataclass
class GeneralVariables(EnvLoader):
    model_bazaar_dir: str
    license_key: str
    model_bazaar_endpoint: str
    model_id: str
    data_id: str
    type: TypeEnum = TypeEnum.NDB


@dataclass
class MachVariables(EnvLoader):
    fhr: int = 50_000
    embedding_dim: int = 2048
    output_dim: int = 10_000
    extreme_num_hashes: int = 1
    hidden_bias: bool = False
    tokenizer: str = "char-4"


@dataclass
class FinetunableRetrieverVariables(EnvLoader):
    on_disk: bool = True


@dataclass
class NeuralDBVariables(EnvLoader):
    num_shards: int = 1
    num_models_per_shard: int = 1
    base_model_id: Optional[str] = None
    retriever: RetrieverEnum = RetrieverEnum.FINETUNABLE_RETRIEVER


@dataclass
class TrainVariables(EnvLoader):
    learning_rate: float = 0.005
    max_in_memory_batches: Optional[int] = None
    batch_size: int = 2048
    unsupervised_epochs: int = 5
    supervised_epochs: int = 3


@dataclass
class S3variables(EnvLoader):
    aws_access_key: str = None
    aws_secret_access_key: str = None
