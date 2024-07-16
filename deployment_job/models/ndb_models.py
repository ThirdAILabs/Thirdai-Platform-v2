import pickle
from abc import abstractmethod

from models.model import Model
from pydantic_models import inputs
from thirdai import neural_db as ndb


class NDBModel(Model):
    def __init__(self):
        super().__init__()
        self.model_path = self.model_dir / "model.ndb"
        self.db: ndb.NeuralDB = self.load_ndb()

    @abstractmethod
    def load_ndb(self):
        pass

    def predict(self, **kwargs):
        constraints = kwargs.get("constraints")

        ndb_constraints = {
            key: getattr(ndb, constraints[key]["constraint_type"])(
                **(
                    {
                        key: value
                        for key, value in constraints[key].items()
                        if key != "constraint_type"
                    }
                )
            )
            for key in constraints.keys()
        }
        references = self.db.search(
            query=kwargs["query"],
            top_k=kwargs["top_k"],
            constraints=ndb_constraints,
            rerank=kwargs.get("rerank", False),
            top_k_rerank=kwargs.get("top_k_rerank", 100),
            rerank_threshold=kwargs.get("rerank_threshold", 1.5),
            top_k_threshold=kwargs.get("top_k_threshold", 10),
        )
        pydantic_references = [
            inputs.convert_reference_to_pydantic(ref, kwargs.get("context_radius", 1))
            for ref in references
        ]
        return inputs.SearchResults(
            query_text=kwargs["query"],
            references=pydantic_references,
        )


class SingleNDB(NDBModel):
    def __init__(self):
        super().__init__()

    def load_ndb(self):
        return ndb.NeuralDB.from_checkpoint(self.model_path)


class ShardedNDB(NDBModel):
    def __init__(self):
        super().__init__()

    def load_ndb(self):
        db = ndb.NeuralDB.from_checkpoint(self.model_path)

        for i in range(db._savable_state.model.num_shards):

            models = []

            for j in range(db._savable_state.model.num_models_per_shard):
                model_shard_num = i * db._savable_state.model.num_models_per_shard + j

                mach_model_pkl = (
                    self.model_dir / str(model_shard_num) / "shard_mach_model.pkl"
                )

                with open(mach_model_pkl, "rb") as pkl:
                    mach_model = pickle.load(pkl)

                models.append(mach_model)

            db._savable_state.model.ensembles[i].models = models

        return db
