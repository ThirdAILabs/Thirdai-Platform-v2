import os
import traceback

import thirdai
from models.multiple_mach import MultipleMach
from models.ndb_models import FinetunableRetriever, SingleMach
from models.classification_models import TextClassificationModel, TokenClassificationModel
from models.shard_mach import ShardMach
from reporter import Reporter
from variables import (
    GeneralVariables,
    NDBSubType,
    NeuralDBVariables,
    RetrieverEnum,
    TypeEnum,
    UDTVariables,
    UDTSubType,
)

general_variables = GeneralVariables.load_from_env()


def main():
    reporter = Reporter(api_url=general_variables.model_bazaar_endpoint)
    try:
        if general_variables.type == TypeEnum.NDB:
            ndb_variables = NeuralDBVariables.load_from_env()
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
        elif general_variables.type == TypeEnum.UDT:
            udt_variables = UDTVariables.load_from_env()
            if udt_variables.sub_type == UDTSubType.text:
                model = TextClassificationModel()
                model.train()
            elif udt_variables.sub_type == UDTSubType.token:
                model = TokenClassificationModel()
                model.train()
            else:
                raise ValueError("Currently Not supported")
            
    except Exception as err:
        traceback.print_exc()
        reporter.report_status(general_variables.model_id, "failed", message=err)


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
