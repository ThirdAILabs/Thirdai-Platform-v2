import os
import queue
import threading
from typing import List

from exeptional_handler import apply_exception_handler
from models.ndb_model_interface import NDBModel
from thirdai import neural_db as ndb
from utils import check_disk, consumer, list_files, producer


@apply_exception_handler
class FinetunableRetriever(NDBModel):
    report_failure_method = "report_status"

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
