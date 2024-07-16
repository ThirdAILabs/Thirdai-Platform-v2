from models.ndb_models import ShardedNDB, SingleNDB
from variables import GeneralVariables, TypeEnum

# Initialize the model to None
model_instance = None


def get_model():
    global model_instance
    if model_instance is None:
        general_variables = GeneralVariables.load_from_env()
        if general_variables.type == TypeEnum.NDB:
            if general_variables.num_shards:
                model_instance = ShardedNDB()
            else:
                model_instance = SingleNDB()
        else:
            raise ValueError("Invalid model type")
    return model_instance
