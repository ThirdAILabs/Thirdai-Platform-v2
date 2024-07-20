import os

import thirdai
from models.finetunable_retriever import FinetunableRetriever
from models.multiple_mach import MultipleMach
from models.shard_mach import ShardMach
from models.single_mach import SingleMach
from variables import (
    GeneralVariables,
    NDBSubType,
    NeuralDBVariables,
    RetrieverEnum,
    TypeEnum,
)

# Load general variables from environment
general_variables: GeneralVariables = GeneralVariables.load_from_env()


def main():
    """
    Main function to initialize and train the appropriate model based on environment variables.
    """
    if general_variables.type == TypeEnum.NDB:
        ndb_variables: NeuralDBVariables = NeuralDBVariables.load_from_env()
        if general_variables.sub_type == NDBSubType.normal:
            if ndb_variables.retriever == RetrieverEnum.FINETUNABLE_RETRIEVER:
                model = FinetunableRetriever()
                model.train()
            else:
                model = SingleMach()
                model.train()
        elif general_variables.sub_type == NDBSubType.shard_allocation:
            if ndb_variables.retriever == RetrieverEnum.FINETUNABLE_RETRIEVER:
                raise ValueError("Currently Not supported")
            else:
                model = MultipleMach()
                model.train()
        else:
            model = ShardMach()
            model.train()


if __name__ == "__main__":
    # Set the license for ThirdAI based on the environment variable
    if general_variables.license_key == "file_license":
        thirdai.licensing.set_path(
            os.path.join(
                general_variables.model_bazaar_dir, "license/license.serialized"
            )
        )
    else:
        thirdai.licensing.activate(general_variables.license_key)

    # Run the main function
    main()
