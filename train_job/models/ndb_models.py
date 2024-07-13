import json
import os
import queue
import threading
from pathlib import Path
from typing import List

import thirdai
from models.model import Model
from thirdai import neural_db as ndb
from utils import (
    check_disk,
    consumer,
    convert_supervised_to_ndb_file,
    get_directory_size,
    list_files,
    producer,
)
from variables import MachVariables, NeuralDBVariables


class NDBModel(Model):
    def __init__(self):
        super().__init__()
        self.ndb_variables = NeuralDBVariables.load_from_env()
        self.model_save_path = self.model_dir / "model.ndb"

    def unsupervised_train(self, db: ndb.NeuralDB, files: List):
        # Look into how to add checkpointing for streaming processes.
        buffer = queue.Queue()

        producer_thread = threading.Thread(
            target=producer,
            args=(files, buffer, self.data_dir / "unsupervised"),
        )

        consumer_thread = threading.Thread(
            target=consumer,
            args=(buffer, db, self.train_variables.unsupervised_epochs, 50),
        )

        producer_thread.start()
        consumer_thread.start()

        producer_thread.join()
        buffer.put(None)  # Signal the consumer to exit
        consumer_thread.join()

    def get_supervised_files(self, files):
        relations_path = self.data_dir / "relations.json"
        if relations_path.exists():
            with relations_path.open("r") as file:
                relations_data = json.load(file)
            relations_dict = {
                entry["supervised_file"]: entry["source_id"] for entry in relations_data
            }

        supervised_source_ids = [relations_dict[Path(file).name] for file in files]

        return [
            convert_supervised_to_ndb_file(file, supervised_source_ids[i])
            for i, file in enumerate(files)
        ]

    def supervised_train(self, db: ndb.NeuralDB, files: List):
        # Look into how to add streaming in this supervised.
        supervised_sources = self.get_supervised_files(files)

        db.supervised_train(
            supervised_sources,
            epochs=self.train_variables.supervised_epochs,
        )

    def get_ndb_path(self, model_id):
        return (
            Path(self.general_variables.model_bazaar_dir)
            / "models"
            / model_id
            / "model.ndb"
        )

    def load_db(self, model_id):
        return ndb.NeuralDB.from_checkpoint(self.get_ndb_path(model_id))

    def get_db(self):
        if self.ndb_variables.base_model_id:
            return self.load_db(self.ndb_variables.base_model_id)
        return self.initialize_db()

    def initialize_db(self):
        pass

    def get_num_params(self, db):
        pass

    def get_size_in_memory(self):
        pass

    def save(self, db):
        db.save(self.model_save_path)

    def finalize_training(self, db):
        num_params = self.get_num_params(db)
        self.save(db)

        size = get_directory_size(self.model_save_path)
        size_in_memory = self.get_size_in_memory()

        # look into adding all the items commented in schema.py file in model table.
        self.reporter.report_complete(
            model_id=self.general_variables.model_id,
            metadata={
                "num_params": str(num_params),
                "size": str(size),
                "size_in_memory": str(size_in_memory),
                "thirdai_version": str(thirdai.__version__),
            },
        )


class SingleMach(NDBModel):
    def __init__(self):
        super().__init__()
        self.mach_variables = MachVariables.load_from_env()

    def train(self, **kwargs):
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        supervised_files = list_files(self.data_dir / "supervised")
        test_files = list_files(self.data_dir / "test")

        db = self.get_db()

        if unsupervised_files:
            check_disk(db, self.general_variables.model_bazaar_dir, unsupervised_files)
            self.unsupervised_train(db, unsupervised_files)
            print("Completed Unsupervised Training", flush=True)
            if test_files:
                self.evaluate(db, test_files)

        if supervised_files:
            check_disk(db, self.general_variables.model_bazaar_dir, supervised_files)
            self.supervised_train(db, supervised_files)
            print("Completed Supervised Training", flush=True)

            if test_files:
                self.evaluate(db, test_files)

        self.finalize_training(db)

    def evaluate(self, db, files, **kwargs):
        for file in files:
            metrics = db._savable_state.model.model.evaluate(
                file,
                metrics=["precision@1", "precision@5", "recall@1", "recall@5"],
            )

    def initialize_db(self):
        return ndb.NeuralDB(
            fhr=self.mach_variables.fhr,
            embedding_dimension=self.mach_variables.embedding_dim,
            extreme_output_dim=self.mach_variables.output_dim,
            extreme_num_hashes=self.mach_variables.extreme_num_hashes,
            tokenizer=self.mach_variables.tokenizer,
            hidden_bias=self.mach_variables.hidden_bias,
            retriever=self.ndb_variables.retriever,
        )

    def get_num_params(self, db):
        model = db._savable_state.model.model._get_model()
        return model.num_params()

    def get_size_in_memory(self):
        udt_pickle = self.model_save_path / "model.pkl"
        documents_pickle = self.model_save_path / "documents.pkl"
        logger_pickle = self.model_save_path / "logger.pkl"

        return (
            os.path.getsize(udt_pickle) * 4
            + os.path.getsize(documents_pickle)
            + os.path.getsize(logger_pickle)
        )


class FinetunableRetriever(NDBModel):
    def __init__(self):
        super().__init__()

    def train(self, **kwargs):
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        supervised_files = list_files(self.data_dir / "supervised")

        db = self.get_db()

        if unsupervised_files:
            check_disk(db, self.general_variables.model_bazaar_dir, unsupervised_files)
            self.unsupervised_train(db, unsupervised_files)
            print("Completed Unsupervised Training", flush=True)

        if supervised_files:
            check_disk(db, self.general_variables.model_bazaar_dir, supervised_files)
            self.supervised_train(db, supervised_files)
            print("Completed Supervised Training", flush=True)

        self.finalize_training(db)

    def evaluate(self, **kwargs):
        pass

    def initialize_db(self):
        return ndb.NeuralDB(
            retriever=self.ndb_variables.retriever,
            on_disk=True,  # See if we have to configure this from variables
        )

    def get_num_params(self, db):
        return sum(doc.size for doc in db._savable_state.documents.sources().values())

    def get_size_in_memory(self):
        udt_pickle = self.model_save_path / "model.pkl"
        documents_pickle = self.model_save_path / "documents.pkl"
        logger_pickle = self.model_save_path / "logger.pkl"

        return int(
            os.path.getsize(udt_pickle) * 1.5
            + os.path.getsize(documents_pickle)
            + os.path.getsize(logger_pickle)
        )
