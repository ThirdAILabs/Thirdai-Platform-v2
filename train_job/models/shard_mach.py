import os
import pickle
from pathlib import Path

import thirdai
from models.ndb_models import NDBModel
from thirdai import data
from thirdai.neural_db.documents import DocumentDataSource
from thirdai.neural_db.models.mach import Mach
from thirdai.neural_db.supervised_datasource import SupDataSource
from thirdai.neural_db.trainer.training_progress_manager import (
    TrainingProgressManager as TPM,
)
from utils import no_op
from variables import MachVariables, ShardVariables


class ShardMach(NDBModel):
    def __init__(self):
        super().__init__()
        self.mach_variables = MachVariables.load_from_env()
        self.shard_variables = ShardVariables.load_from_env()

    def get_model_path(self, model_id):
        return (
            Path(self.get_ndb_path(model_id)).parent
            / str(self.shard_variables.shard_num)
            / "shard_mach_model.pkl"
        )

    def get_model(self, data_shard_num, model_num_in_shard):
        if self.ndb_variables.base_model_id:
            base_model_path = self.get_model_path(self.ndb_variables.base_model_id)

            with base_model_path.open("rb") as pkl:
                return pickle.load(pkl)

        return Mach(
            id_col="id",
            id_delimiter=" ",
            query_col="query",
            fhr=self.mach_variables.fhr,
            embedding_dimension=self.mach_variables.embedding_dim,
            extreme_output_dim=self.mach_variables.output_dim,
            extreme_num_hashes=self.mach_variables.extreme_num_hashes,
            mach_index_seed=data_shard_num * 341 + model_num_in_shard * 17,
            hybrid=(
                (self.ndb_variables.retriever == "hybrid")
                if model_num_in_shard == 0
                else False
            ),
            hidden_bias=self.mach_variables.hidden_bias,
            tokenizer=self.mach_variables.tokenizer,
        )

    def get_shard_data(self, path):
        with path.open("rb") as pkl:
            shard_picklable = pickle.load(pkl)
        shard = DocumentDataSource(
            shard_picklable["id_column"],
            shard_picklable["strong_column"],
            shard_picklable["weak_column"],
        )
        shard.documents = shard_picklable["documents"]
        shard._size = shard_picklable["_size"]
        for i, (doc, _) in enumerate(shard.documents):
            doc.path = Path(f"/shard_{self.shard_variables.shard_num}_doc_{i}.shard")
        return shard

    def train(self, **kwargs):
        self.reporter.report_shard_train_status(
            self.general_variables.model_id,
            self.shard_variables.shard_num,
            "in_progress",
        )

        thirdai.logging.setup(
            log_to_stderr=False,
            path=str(
                self.model_dir / f"train-shard-{self.shard_variables.shard_num}.log"
            ),
            level="info",
        )

        data_shard_num = int(
            self.shard_variables.shard_num / self.ndb_variables.num_models_per_shard
        )
        model_num_in_shard = (
            self.shard_variables.shard_num % self.ndb_variables.num_models_per_shard
        )

        mach_model = self.get_model(data_shard_num, model_num_in_shard)

        test_file = self.data_dir / "test" / f"shard_{data_shard_num}.csv"

        if (
            intro_shard_path := self.data_dir / f"intro_shard_{data_shard_num}.pkl"
        ).exists():
            intro_shard = self.get_shard_data(intro_shard_path)
            train_shard = self.get_shard_data(
                self.data_dir / f"train_shard_{data_shard_num}.pkl"
            )

            training_manager = TPM.from_scratch_for_unsupervised(
                model=mach_model,
                intro_documents=intro_shard,
                train_documents=train_shard,
                should_train=self.train_variables.unsupervised_train,
                override_number_classes=self.shard_variables.num_classes,
                variable_length=data.transformations.VariableLengthConfig(),
                fast_approximation=self.train_variables.fast_approximation,
                num_buckets_to_sample=None,
                max_in_memory_batches=self.train_variables.max_in_memory_batches,
                epochs=self.train_variables.unsupervised_epochs,
                learning_rate=self.train_variables.learning_rate,
                batch_size=self.train_variables.batch_size,
                checkpoint_config=None,
            )

            mach_model.index_documents_impl(
                training_progress_manager=training_manager,
                on_progress=no_op,
                cancel_state=None,
            )

            if test_file.exists():
                self.evaluate(mach_model, test_file)

        if (
            supervised_shard_path := self.data_dir
            / f"supervised_shard_{data_shard_num}.pkl"
        ).exists():
            with supervised_shard_path.open("rb") as pkl:
                supervised_shard_picklable = pickle.load(pkl)

            supervised_data_source = SupDataSource(
                query_col=supervised_shard_picklable["query_column"],
                data=supervised_shard_picklable["data"],
                id_delimiter=supervised_shard_picklable["id_delimiter"],
                id_column=supervised_shard_picklable["id_column"],
            )
            del supervised_shard_picklable

            mach_model.train_on_supervised_data_source(
                supervised_data_source,
                learning_rate=self.train_variables.learning_rate,
                epochs=self.train_variables.supervised_epochs,
                max_in_memory_batches=self.train_variables.max_in_memory_batches,
                batch_size=self.train_variables.batch_size,
                metrics=["loss", "hash_precision@1"],
                callbacks=[],
                disable_finetunable_retriever=self.train_variables.disable_finetunable_retriever,
            )

            if test_file.exists():
                self.evaluate(mach_model, test_file)

        shard_model_path = self.get_model_path(self.general_variables.model_id)
        shard_model_path.parent.mkdir(parents=True, exist_ok=True)

        with shard_model_path.open("wb") as pkl:
            pickle.dump(mach_model, pkl)

        print("Saved Mach model", flush=True)

        self.reporter.report_shard_train_status(
            self.general_variables.model_id, self.shard_variables.shard_num, "complete"
        )

    def evaluate(self, model: Mach, file, **kwargs):
        metrics = model.model.evaluate(
            file,
            metrics=["precision@1", "precision@5", "recall@1", "recall@5"],
        )
