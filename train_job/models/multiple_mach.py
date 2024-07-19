import os
import pickle
import time
from pathlib import Path

import thirdai
from models.ndb_models import NDBModel
from thirdai import neural_db as ndb
from thirdai.neural_db.models.mach_mixture_model import MachMixture
from thirdai.neural_db.sharded_documents import shard_data_source
from thirdai.neural_db.supervised_datasource import SupDataSource
from utils import convert_to_ndb_file, list_files, make_test_shard_files
from variables import ComputeVariables, MachVariables, merge_dataclasses_to_dict


class MultipleMach(NDBModel):
    def __init__(self):
        """
        Initialize the MultipleMach model with environment variables.
        """
        super().__init__()
        self.mach_variables: MachVariables = MachVariables.load_from_env()

    def initialize_db(self) -> ndb.NeuralDB:
        """
        Initialize the NeuralDB with the required parameters.
        Returns:
            ndb.NeuralDB: The initialized NeuralDB instance.
        """
        return ndb.NeuralDB(
            fhr=self.mach_variables.fhr,
            embedding_dimension=self.mach_variables.embedding_dim,
            extreme_output_dim=self.mach_variables.output_dim,
            extreme_num_hashes=self.mach_variables.extreme_num_hashes,
            tokenizer=self.mach_variables.tokenizer,
            hidden_bias=self.mach_variables.hidden_bias,
            retriever=self.ndb_variables.retriever,
            num_models_per_shard=self.ndb_variables.num_models_per_shard,
            num_shards=self.ndb_variables.num_shards,
        )

    def get_approx_ndb_size(
        self,
        num_labels: int,
        documents: list[str],
        num_shards: int,
        num_models_per_shard: int,
    ) -> tuple[int, int, int, int]:
        """
        Estimate the approximate size of the NeuralDB.
        Args:
            num_labels (int): Number of labels in the data.
            documents (list[str]): List of document paths.
            num_shards (int): Number of shards.
            num_models_per_shard (int): Number of models per shard.
        Returns:
            tuple[int, int, int, int]: Tuple containing the total NDB size, model size, document size, and total model parameters.
        """
        model_params_each = (
            self.mach_variables.fhr + self.mach_variables.output_dim
        ) * self.mach_variables.embedding_dim  # bolt model params
        model_params_each += (
            2 * num_labels * self.mach_variables.extreme_num_hashes  # mach index size
        )

        model_params_total = model_params_each * num_shards * num_models_per_shard
        total_model_size = model_params_total * 4  # int: 4 bytes
        total_model_size *= 1.25  # approximation

        doc_size = sum(
            2 * os.path.getsize(doc_path) for doc_path in documents
        )  # documents and documents.pkl stored in ndb
        total_ndb_size = total_model_size + doc_size

        return total_ndb_size, total_model_size, doc_size, model_params_total

    def load_db(self) -> ndb.NeuralDB:
        """
        Load the NeuralDB from a checkpoint.
        Returns:
            ndb.NeuralDB: The loaded NeuralDB instance.
        """
        db = ndb.NeuralDB.from_checkpoint(
            self.get_ndb_path(self.ndb_variables.base_model_id)
        )

        mixture: MachMixture = db._savable_state.model
        if self.ndb_variables.num_shards != mixture.num_shards:
            raise Exception(
                f"Number of Shards in the base model is {mixture.num_shards} which is not equal to the value for num_shards specified in the argument"
            )

        if self.ndb_variables.num_models_per_shard != mixture.num_models_per_shard:
            raise Exception(
                f"Number of Models per shard in the base model is {mixture.num_models_per_shard} which is not equal to the value for num_models_per_shard specified in the argument"
            )

        return db

    def get_data_shard_dict(self, data_shard) -> dict:
        """
        Get the dictionary representation of a data shard.
        Args:
            data_shard: The data shard to convert.
        Returns:
            dict: Dictionary representation of the data shard.
        """
        return {
            "documents": data_shard.documents,
            "id_column": data_shard.id_column,
            "strong_column": data_shard.strong_column,
            "weak_column": data_shard.weak_column,
            "_size": data_shard._size,
        }

    def save_data_shards(self, intro_shards, train_shards):
        """
        Save the data shards to disk.
        Args:
            intro_shards: List of introduction data shards.
            train_shards: List of training data shards.
        """
        for i, (intro_shard, train_shard) in enumerate(zip(intro_shards, train_shards)):
            with (self.data_dir / f"intro_shard_{i}.pkl").open("wb") as pkl:
                pickle.dump(self.get_data_shard_dict(intro_shard), pkl)
            with (self.data_dir / f"train_shard_{i}.pkl").open("wb") as pkl:
                pickle.dump(self.get_data_shard_dict(train_shard), pkl)

    def save_supervised_shards(self, supervised_shards):
        """
        Save the supervised data shards to disk.
        Args:
            supervised_shards: List of supervised data shards.
        """
        for i in range(len(supervised_shards)):
            supervised_shard_picklable = {
                "id_column": supervised_shards[i].id_column,
                "query_column": supervised_shards[i].query_col,
                "data": supervised_shards[i].data,
                "id_delimiter": supervised_shards[i].id_delimiter,
            }

            with (self.data_dir / f"supervised_shard_{i}.pkl").open("wb") as pkl:
                pickle.dump(supervised_shard_picklable, pkl)

    def save(self, db: ndb.NeuralDB):
        """
        Save the NeuralDB to disk.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance to save.
        """
        for ensemble in db._savable_state.model.ensembles:
            ensemble.set_model([])
        db.save(self.model_save_path)

    def create_extra_options(self) -> dict:
        """
        Create extra options for shard creation.
        Returns:
            dict: Dictionary of extra options.
        """
        compute_variables = ComputeVariables.load_from_env()
        extra_options = merge_dataclasses_to_dict(
            self.mach_variables,
            self.ndb_variables,
            compute_variables,
            self.train_variables,
        )
        extra_options["allocation_memory"] = extra_options["model_memory"]
        extra_options["allocation_cores"] = extra_options["model_cores"]
        extra_options.pop("base_model_id")

        return extra_options

    def train(self, **kwargs):
        """
        Train the MultipleMach model.
        """
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        db = self.get_db()

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        supervised_files = list_files(self.data_dir / "supervised")
        test_files = list_files(self.data_dir / "test")

        mixture: MachMixture = db._savable_state.model

        doc_manager = db._savable_state.documents
        total_num_labels = doc_manager.get_data_source().size
        documents = doc_manager.sources()
        total_documents = [value.path for _, value in documents.items()]

        if unsupervised_files:
            unsupervised_sources = [
                convert_to_ndb_file(file) for file in unsupervised_files
            ]

            intro_and_train, _ = db._savable_state.documents.add(unsupervised_sources)

            total_num_labels += intro_and_train.intro.size
            total_documents.extend(unsupervised_files)

            introduce_data_sources = shard_data_source(
                data_source=intro_and_train.intro,
                label_to_segment_map=mixture.label_to_segment_map,
                number_shards=mixture.num_shards,
                update_segment_map=True,
            )

            train_data_sources = shard_data_source(
                data_source=intro_and_train.train,
                label_to_segment_map=mixture.label_to_segment_map,
                number_shards=mixture.num_shards,
                update_segment_map=False,
            )

            self.save_data_shards(introduce_data_sources, train_data_sources)

        if supervised_files:
            supervised_sources = self.get_supervised_files(supervised_files)
            doc_manager = db._savable_state.documents
            supervised_data_source = SupDataSource(
                doc_manager=doc_manager,
                query_col=db._savable_state.model.get_query_col(),
                data=supervised_sources,
                id_delimiter=db._savable_state.model.get_id_delimiter(),
            )

            sharded_supervised_datasource = shard_data_source(
                data_source=supervised_data_source,
                number_shards=mixture.num_shards,
                label_to_segment_map=mixture.label_to_segment_map,
                update_segment_map=False,
            )

            self.save_supervised_shards(sharded_supervised_datasource)

        if test_files:
            make_test_shard_files(
                test_files[0],
                mixture.label_to_segment_map,
                self.data_dir / "test",
                mixture.get_id_col(),
                mixture.get_id_delimiter(),
            )

        approx_ndb_size, model_size, doc_size, model_params = self.get_approx_ndb_size(
            num_labels=total_num_labels,
            documents=total_documents,
            num_shards=mixture.num_shards,
            num_models_per_shard=mixture.num_models_per_shard,
        )

        self.save(db)
        extra_options = self.create_extra_options()
        extra_options["num_classes"] = total_num_labels
        for i in range(mixture.num_shards):
            for j in range(mixture.num_models_per_shard):
                self.reporter.create_shard(
                    shard_num=i * mixture.num_models_per_shard + j,
                    model_id=self.general_variables.model_id,
                    data_id=self.general_variables.data_id,
                    base_model_id=self.ndb_variables.base_model_id,
                    extra_options=extra_options,
                )

        all_shard_status = False
        while not all_shard_status:
            content = self.reporter.get_model_shard_train_status(
                self.general_variables.model_id
            )
            model_shard_train_status = content.get("data")

            if (
                len(model_shard_train_status)
                != mixture.num_shards * mixture.num_models_per_shard
            ):
                raise Exception(
                    "Number of shards training is not equal to actual number of shards"
                )

            all_shard_status = all(
                [shard["status"] == "complete" for shard in model_shard_train_status]
            )

            fail_shard_status = any(
                [shard["status"] == "failed" for shard in model_shard_train_status]
            )
            if fail_shard_status:
                raise Exception("model shard train failure")

            time.sleep(10)

        self.reporter.report_complete(
            self.general_variables.model_id,
            metadata={
                "num_params": str(model_params),
                "size": str(approx_ndb_size),
                "size_in_memory": str(int(model_size * 4 + doc_size)),
                "thirdai_version": str(thirdai.__version__),
            },
        )

    def evaluate(self, **kwargs):
        """
        Evaluate the MultipleMach model.
        """
        pass
