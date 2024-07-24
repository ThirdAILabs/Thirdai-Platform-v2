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

# Initialize the model to None
model_instance = None
general_variables: GeneralVariables = GeneralVariables.load_from_env()

if general_variables.license_key == "file_license":
    thirdai.licensing.set_path(
        os.path.join(general_variables.model_bazaar_dir, "license/license.serialized")
    )
else:
    thirdai.licensing.activate(general_variables.license_key)


def get_model():
    """
    Retrieves the appropriate model instance based on general variables.

    Returns:
        Union[ShardedNDB, SingleNDB]: The initialized model instance.

    Raises:
        ValueError: If the model type is invalid.
    """
    global model_instance
    if model_instance is None:
        if general_variables.type == TypeEnum.NDB:
            if general_variables.num_shards:
                model_instance = ShardedNDB()
            else:
                model_instance = SingleNDB()
        elif general_variables.type == TypeEnum.UDT:
            if general_variables.sub_type == UDTSubtype.text:
                model_instance = TextClassificationModel()
            else:
                model_instance = TokenClassificationModel()
        else:
            raise ValueError("Invalid model type")
    return model_instance
