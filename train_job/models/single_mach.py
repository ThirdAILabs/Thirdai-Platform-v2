import os
import shutil
import time
from typing import List

from exceptional_handler import apply_exception_handler
from models.ndb_model_interface import NDBModel
from thirdai import neural_db as ndb
from utils import check_disk, list_files, process_file
from variables import MachVariables


@apply_exception_handler
class SingleMach(NDBModel):
    report_failure_method = "report_status"

    def __init__(self):
        """
        Initialize the SingleMach model with general, NeuralDB-specific, and Mach-specific variables.
        """
        super().__init__()
        self.mach_variables: MachVariables = MachVariables.load_from_env()
        self.logger.info("SingleMach initialized with Mach-specific variables.")

    def unsupervised_train(self, db: ndb.NeuralDB, files: List[str]):
        """
        Train the model with unsupervised data.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for unsupervised training data.
        """
        self.logger.info("Starting unsupervised training.")
        unsupervised_docs = [
            process_file(file, self.data_dir / "unsupervised") for file in files
        ]
        self.logger.info(f"Processed {len(unsupervised_docs)} unsupervised documents.")

        db.insert(
            unsupervised_docs,
            train=True,
            checkpoint_config=self.unsupervised_checkpoint_config,
            epochs=self.train_variables.unsupervised_epochs,
        )
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
            checkpoint_config=self.supervised_checkpoint_config,
        )
        self.logger.info("Completed supervised training.")

    def train(self, **kwargs):
        """
        Train the SingleMach model with unsupervised and supervised data.
        """
        self.logger.info("Training process started.")
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        supervised_files = list_files(self.data_dir / "supervised")
        test_files = list_files(self.data_dir / "test")

        db = self.get_db()

        start_time = time.time()

        if unsupervised_files:
            self.logger.info(f"Found {len(unsupervised_files)} unsupervised files.")
            check_disk(db, self.general_variables.model_bazaar_dir, unsupervised_files)
            self.unsupervised_train(db, unsupervised_files)
            self.logger.info("Completed Unsupervised Training")
            if test_files:
                self.evaluate(db, test_files)

        if supervised_files:
            self.logger.info(f"Found {len(supervised_files)} supervised files.")
            check_disk(db, self.general_variables.model_bazaar_dir, supervised_files)
            self.supervised_train(db, supervised_files)
            self.logger.info("Completed Supervised Training")

            if test_files:
                self.evaluate(db, test_files)

        total_time = time.time() - start_time
        self.logger.info(f"Total training time: {total_time} seconds")

        self.save(db)
        self.logger.info("Model saved successfully.")

        if self.unsupervised_checkpoint_dir.exists():
            shutil.rmtree(self.unsupervised_checkpoint_dir)
            self.logger.info(
                f"Removed unsupervised checkpoint directory: {self.unsupervised_checkpoint_dir}"
            )
        if self.supervised_checkpoint_dir.exists():
            shutil.rmtree(self.supervised_checkpoint_dir)
            self.logger.info(
                f"Removed supervised checkpoint directory: {self.supervised_checkpoint_dir}"
            )

        self.finalize_training(db, total_time)
        self.logger.info("Training finalized successfully.")

    def evaluate(self, db: ndb.NeuralDB, files: List[str], **kwargs):
        """
        Evaluate the model with the given test files.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
            files (List[str]): List of file paths for evaluation data.
        """
        self.logger.info("Starting evaluation process.")
        for file in files:
            metrics = db._savable_state.model.model.evaluate(
                file, metrics=self.train_variables.metrics
            )
            self.logger.info(f"For file {file} metrics are {metrics}")

    def initialize_db(self) -> ndb.NeuralDB:
        """
        Initialize a new NeuralDB instance with the required parameters.
        Returns:
            ndb.NeuralDB: The initialized NeuralDB instance.
        """
        self.logger.info("Initializing a new NeuralDB instance.")
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
        num_params = model.num_params()
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

        size_in_memory = (
            os.path.getsize(udt_pickle) * 4
            + os.path.getsize(documents_pickle)
            + os.path.getsize(logger_pickle)
        )
        self.logger.info(f"Size of the model in memory: {size_in_memory} bytes")
        return size_in_memory
