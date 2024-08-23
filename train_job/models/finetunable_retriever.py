import os
import queue
import threading
import time
from typing import List

from exceptional_handler import apply_exception_handler
from models.ndb_model_interface import NDBModel
from thirdai import neural_db as ndb
from utils import check_disk, consumer, list_files, producer
from variables import FinetunableRetrieverVariables


@apply_exception_handler
class FinetunableRetriever(NDBModel):
    report_failure_method = "report_status"

    def __init__(self):
        """
        Initialize the FinetunableRetriever model with general and NeuralDB-specific variables.
        """
        super().__init__()
        self.finetunable_retriever_variables = (
            FinetunableRetrieverVariables.load_from_env()
        )
        self.logger.info("FinetunableRetriever initialized with variables.")

    def unsupervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with unsupervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for unsupervised training data.
        """
        self.logger.info("Starting unsupervised training.")
        buffer = queue.Queue()

        producer_thread = threading.Thread(
            target=producer,
            args=(files, buffer, self.data_dir / "unsupervised"),
        )

        consumer_thread = threading.Thread(
            target=consumer,
            args=(buffer, db, self.train_variables.unsupervised_epochs, 50),
        )

        self.logger.info(
            "Starting producer and consumer threads for unsupervised training."
        )
        producer_thread.start()
        consumer_thread.start()

        producer_thread.join()
        buffer.put(None)  # Signal the consumer to exit
        consumer_thread.join()
        self.logger.info("Completed unsupervised training.")

    def supervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with supervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for supervised training data.
        """
        self.logger.info("Starting supervised training.")
        supervised_sources = self.get_supervised_files(files)

        db.supervised_train(
            supervised_sources,
            epochs=self.train_variables.supervised_epochs,
        )
        self.logger.info("Completed supervised training.")

    def train(self, **kwargs):
        """
        Train the FinetunableRetriever model with unsupervised and supervised data.
        """
        self.logger.info("Training process started.")
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        supervised_files = list_files(self.data_dir / "supervised")

        db = self.get_db()

        start_time = time.time()

        if unsupervised_files:
            self.logger.info(f"Found {len(unsupervised_files)} unsupervised files.")
            check_disk(db, self.general_variables.model_bazaar_dir, unsupervised_files)
            self.unsupervised_train(db, unsupervised_files)
            self.logger.info("Completed Unsupervised Training")

        if supervised_files:
            self.logger.info(f"Found {len(supervised_files)} supervised files.")
            check_disk(db, self.general_variables.model_bazaar_dir, supervised_files)
            self.supervised_train(db, supervised_files)
            self.logger.info("Completed Supervised Training")

        total_time = time.time() - start_time
        self.logger.info(f"Total training time: {total_time} seconds")

        self.save(db)
        self.logger.info("Model saved successfully.")

        self.finalize_training(db, total_time)
        self.logger.info("Training finalized successfully.")

    def evaluate(self, **kwargs):
        """
        Evaluate the FinetunableRetriever model. Not implemented.
        """
        self.logger.warning("Evaluation method called. Not implemented.")

    def load_db(self, model_id: str) -> ndb.NeuralDB:
        """
        This method first loads the base NeuralDB using the parent class's `load_db` method.
        Since we use an on-disk inverted index, we need to ensure the base NeuralDB is saved
        in the specified `model_save_path` before making any modifications to it.
        Steps:
        1. Load the base NeuralDB using the parent method.
        2. Save the base NeuralDB to the specified `model_save_path`.
        3. Reload the NeuralDB from the saved checkpoint to ensure modifications are performed
        on the correct instance and it doesn't affect the base model index.
        """
        db = super().load_db(model_id)
        db.save(self.model_save_path)
        return ndb.NeuralDB.from_checkpoint(self.model_save_path)

    def save(self, db: ndb.NeuralDB):
        """
        This method checks if the current `model_save_path` already exists. If it does, it implies
        that the current training started from a pre-existing base model that was saved earlier,
        so we skip the save operation to avoid redundant saving. Otherwise, it saves the current
        NeuralDB using the parent class's `save` method.
        """
        if not self.model_save_path.exists():
            super().save(db)

    def initialize_db(self) -> ndb.NeuralDB:
        """
        Initialize a new NeuralDB instance with the retriever.
        Returns:
            ndb.NeuralDB: The initialized NeuralDB instance.
        """
        self.logger.info("Initializing a new NeuralDB instance.")
        return ndb.NeuralDB(
            retriever=self.ndb_variables.retriever,
            on_disk=self.finetunable_retriever_variables.on_disk,
        )

    def get_num_params(self, db: ndb.NeuralDB) -> int:
        """
        Get the number of parameters in the model.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
        Returns:
            int: The number of parameters in the model.
        """
        num_params = sum(
            doc.size for doc in db._savable_state.documents.sources().values()
        )
        self.logger.info(f"Number of parameters in the model: {num_params}")
        return num_params

    def get_size_in_memory(self) -> int:
        """
        Get the size of the model in memory.
        Returns:
            int: The size of the model in memory.
        """
        udt_pickle = self.model_save_path / "model.pkl"
        documents_pickle = self.model_save_path / "documents.pkl"
        logger_pickle = self.model_save_path / "logger.pkl"

        size_in_memory = int(
            os.path.getsize(udt_pickle) * 1.5
            + os.path.getsize(documents_pickle)
            + os.path.getsize(logger_pickle)
        )
        self.logger.info(f"Size of the model in memory: {size_in_memory} bytes")
        return size_in_memory
