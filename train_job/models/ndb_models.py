import json
import os
import queue
import shutil
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
    process_file,
    producer,
)
from variables import MachVariables, NeuralDBVariables


class NDBModel(Model):
    def __init__(self):
        """
        Initialize the NDBModel with general and NeuralDB-specific variables.
        """
        super().__init__()
        self.ndb_variables: NeuralDBVariables = NeuralDBVariables.load_from_env()
        self.model_save_path: Path = self.model_dir / "model.ndb"

        self.unsupervised_checkpoint_config = (
            ndb.CheckpointConfig(
                checkpoint_dir=self.get_checkpoint_dir(
                    self.unsupervised_checkpoint_dir
                ),
                checkpoint_interval=self.train_variables.checkpoint_interval,
                resume_from_checkpoint=self.get_checkpoint_dir(
                    self.unsupervised_checkpoint_dir
                ).exists(),
            )
            if self.train_variables.checkpoint_interval
            else None
        )

        self.supervised_checkpoint_config = (
            ndb.CheckpointConfig(
                checkpoint_dir=self.get_checkpoint_dir(self.supervised_checkpoint_dir),
                checkpoint_interval=self.train_variables.checkpoint_interval,
                resume_from_checkpoint=self.get_checkpoint_dir(
                    self.supervised_checkpoint_dir
                ).exists(),
            )
            if self.train_variables.checkpoint_interval
            else None
        )

    def get_checkpoint_dir(self, base_checkpoint_dir: Path):
        return base_checkpoint_dir

    def get_supervised_files(self, files: List[str]) -> List[ndb.Sup]:
        """
        Convert files to supervised NDB files.
        Args:
            files (List[str]): List of file paths for supervised training data.
        Returns:
            List[ndb.Sup]: List of converted supervised NDB files.
        """
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

    def get_ndb_path(self, model_id: str) -> Path:
        """
        Get the path to the NeuralDB checkpoint for a given model ID.
        Args:
            model_id (str): The model ID.
        Returns:
            Path: The path to the NeuralDB checkpoint.
        """
        return (
            Path(self.general_variables.model_bazaar_dir)
            / "models"
            / model_id
            / "model.ndb"
        )

    def load_db(self, model_id: str) -> ndb.NeuralDB:
        """
        Load the NeuralDB from a checkpoint.
        Args:
            model_id (str): The model ID.
        Returns:
            ndb.NeuralDB: The loaded NeuralDB instance.
        """
        return ndb.NeuralDB.from_checkpoint(self.get_ndb_path(model_id))

    def get_db(self) -> ndb.NeuralDB:
        """
        Get the NeuralDB instance, either by loading from a checkpoint or initializing a new one.
        Returns:
            ndb.NeuralDB: The NeuralDB instance.
        """
        if self.ndb_variables.base_model_id:
            return self.load_db(self.ndb_variables.base_model_id)
        return self.initialize_db()

    def initialize_db(self):
        """
        Initialize a new NeuralDB instance. Must be implemented by subclasses with single training.
        """
        pass

    def get_num_params(self, db: ndb.NeuralDB):
        """
        Get the number of parameters in the model. Must be implemented by subclasses with single training.
        """
        pass

    def get_size_in_memory(self) -> int:
        """
        Get the size of the model in memory. Must be implemented by subclasses with single training
        """
        pass

    def save(self, db: ndb.NeuralDB):
        """
        Save the NeuralDB to disk.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance to save.
        """
        db.save(self.model_save_path)

    def finalize_training(self, db: ndb.NeuralDB):
        """
        Finalize the training process by saving the model and reporting completion.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
        """
        num_params = self.get_num_params(db)

        size = get_directory_size(self.model_save_path)
        size_in_memory = self.get_size_in_memory()

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
        """
        Initialize the SingleMach model with general, NeuralDB-specific, and Mach-specific variables.
        """
        super().__init__()
        self.mach_variables: MachVariables = MachVariables.load_from_env()

    def unsupervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with unsupervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for unsupervised training data.
        """
        # For mach we need to have all the files in insert otherwise mach has
        # this forgetting nature, so not doing the streaming way for mach.
        unsupervised_docs = [
            process_file(file, self.data_dir / "unsupervised") for file in files
        ]

        db.insert(
            unsupervised_docs,
            train=True,
            checkpoint_config=self.unsupervised_checkpoint_config,
            epochs=self.train_variables.unsupervised_epochs,
        )

    def supervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with supervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for supervised training data.
        """
        supervised_sources = self.get_supervised_files(files)

        db.supervised_train(
            supervised_sources,
            epochs=self.train_variables.supervised_epochs,
            checkpoint_config=self.supervised_checkpoint_config,
        )

    def train(self, **kwargs):
        """
        Train the SingleMach model with unsupervised and supervised data.
        """
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

        self.save(db)

        if self.unsupervised_checkpoint_dir.exists():
            shutil.rmtree(self.unsupervised_checkpoint_dir)
        if self.supervised_checkpoint_dir.exists():
            shutil.rmtree(self.supervised_checkpoint_dir)

        self.finalize_training(db)

    def evaluate(self, db: ndb.NeuralDB, files: List[str], **kwargs):
        """
        Evaluate the model with the given test files.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for evaluation data.
        """
        for file in files:
            metrics = db._savable_state.model.model.evaluate(
                file,
                metrics=["precision@1", "precision@5", "recall@1", "recall@5"],
            )

    def initialize_db(self) -> ndb.NeuralDB:
        """
        Initialize a new NeuralDB instance with the required parameters.
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
        )

    def get_num_params(self, db: ndb.NeuralDB) -> int:
        """
        Get the number of parameters in the model.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
        Returns:
            int: The number of parameters in the model.
        """
        model = db._savable_state.model.model._get_model()
        return model.num_params()

    def get_size_in_memory(self) -> int:
        """
        Get the size of the model in memory.
        Returns:
            int: The size of the model in memory.
        """
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
        """
        Initialize the FinetunableRetriever model with general and NeuralDB-specific variables.
        """
        super().__init__()

    def unsupervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with unsupervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for unsupervised training data.
        """
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

    def supervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with supervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for supervised training data.
        """
        supervised_sources = self.get_supervised_files(files)

        db.supervised_train(
            supervised_sources,
            epochs=self.train_variables.supervised_epochs,
        )

    def train(self, **kwargs):
        """
        Train the FinetunableRetriever model with unsupervised and supervised data.
        """
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

        self.save(db)

        self.finalize_training(db)

    def evaluate(self, **kwargs):
        """
        Evaluate the FinetunableRetriever model. Not implemented.
        """
        pass

    def initialize_db(self) -> ndb.NeuralDB:
        """
        Initialize a new NeuralDB instance with the retriever.
        Returns:
            ndb.NeuralDB: The initialized NeuralDB instance.
        """
        return ndb.NeuralDB(
            retriever=self.ndb_variables.retriever,
            on_disk=True,  # See if we have to configure this from variables
        )

    def get_num_params(self, db: ndb.NeuralDB) -> int:
        """
        Get the number of parameters in the model.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
        Returns:
            int: The number of parameters in the model.
        """
        return sum(doc.size for doc in db._savable_state.documents.sources().values())

    def get_size_in_memory(self) -> int:
        """
        Get the size of the model in memory.
        Returns:
            int: The size of the model in memory.
        """
        udt_pickle = self.model_save_path / "model.pkl"
        documents_pickle = self.model_save_path / "documents.pkl"
        logger_pickle = self.model_save_path / "logger.pkl"

        return int(
            os.path.getsize(udt_pickle) * 1.5
            + os.path.getsize(documents_pickle)
            + os.path.getsize(logger_pickle)
        )
