"""
Main module to initialize and retrieve the appropriate model instance.
"""

import os

import thirdai
from models.classification_models import (
    TextClassificationModel,
    TokenClassificationModel,
)
from models.ndb_models import ShardedNDB, SingleNDB
from variables import GeneralVariables, NDBSubtype, TypeEnum, UDTSubtype

general_variables: GeneralVariables = GeneralVariables.load_from_env()

if general_variables.license_key == "file_license":
    thirdai.licensing.set_path(
        os.path.join(general_variables.model_bazaar_dir, "license/license.serialized")
    )
else:
    thirdai.licensing.activate(general_variables.license_key)


# Singleton Practice for Model instances.
class ModelManager:
    _model_instance = None

    @classmethod
    def get_instance(cls, write_mode: bool = False):
        """
        Retrieves the appropriate model instance based on general variables.

        Returns:
            The initialized model instance.

        Raises:
            ValueError: If the model type is invalid.
        """
        if cls._model_instance is None:
            if general_variables.type == TypeEnum.NDB:
                if general_variables.sub_type == NDBSubtype.sharded:
                    cls._model_instance = ShardedNDB(write_mode=write_mode)
                else:
                    cls._model_instance = SingleNDB(write_mode=write_mode)
            elif general_variables.type == TypeEnum.UDT:
                if general_variables.sub_type == UDTSubtype.text:
                    cls._model_instance = TextClassificationModel()
                else:
                    cls._model_instance = TokenClassificationModel()
            else:
                raise ValueError("Invalid model type")
        return cls._model_instance


def get_model(write_mode: bool = False):
    return ModelManager.get_instance(write_mode=write_mode)
