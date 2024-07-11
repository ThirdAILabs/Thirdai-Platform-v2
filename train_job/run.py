import os

import thirdai
from model import FinetunableRetriever, SingleMach
from variables import GeneralVariables, NeuralDBVariables, RetrieverEnum, TypeEnum

general_variables = GeneralVariables.load_from_env()


def main():
    try:
        if general_variables.type == TypeEnum.NDB:
            ndb_variables = NeuralDBVariables.load_from_env()
            if ndb_variables.retriever == RetrieverEnum.FINETUNABLE_RETRIEVER:
                model = FinetunableRetriever()
                model.train()
            elif ndb_variables.num_models_per_shard > 1 or ndb_variables.num_shards > 1:
                # Add the class for sharded training.
                pass
            else:
                model = SingleMach()
                model.train()
    except:
        raise


if __name__ == "__main__":
    if general_variables.license_key == "file_license":
        thirdai.licensing.set_path(
            os.path.join(
                general_variables.model_bazaar_dir, "license/license.serialized"
            )
        )
    else:
        thirdai.licensing.activate(general_variables.license_key)

    main()
