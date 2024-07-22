import json
import time
from pathlib import Path
from typing import List

import thirdai
from exeptional_handler import apply_exception_handler
from models.model import Model
from thirdai import neural_db as ndb
from utils import convert_supervised_to_ndb_file, get_directory_size
from variables import NeuralDBVariables


@apply_exception_handler
class NDBModel(Model):
    report_failure_method = "report_status"

    def __init__(self):
        """
        Initialize the NDBModel with general and NeuralDB-specific variables.
        """
        super().__init__()
        self.ndb_variables: NeuralDBVariables = NeuralDBVariables.load_from_env()
        self.model_save_path: Path = self.model_dir / "model.ndb"

        self.unsupervised_checkpoint_config = self.create_checkpoint_config(
            self.get_checkpoint_dir(self.unsupervised_checkpoint_dir)
        )

        self.supervised_checkpoint_config = self.create_checkpoint_config(
            self.get_checkpoint_dir(self.supervised_checkpoint_dir)
        )

    def create_checkpoint_config(self, dir_path: Path):
        return (
            ndb.CheckpointConfig(
                checkpoint_dir=dir_path,
                checkpoint_interval=self.train_variables.checkpoint_interval,
                resume_from_checkpoint=dir_path.exists(),
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

    def get_latency(self, db: ndb.NeuralDB) -> float:
        """
        Get the latency of the db, Must be implemented by subclasses with single training
        """
        start_time = time.time()

        db.search("Checking for latency", top_k=5)

        return time.time() - start_time

    def save(self, db: ndb.NeuralDB):
        """
        Save the NeuralDB to disk.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance to save.
        """
        db.save(self.model_save_path)

    def finalize_training(self, db: ndb.NeuralDB, train_time: int):
        """
        Finalize the training process by saving the model and reporting completion.
        Args:
            db (ndb.NeuralDB): The NeuralDB instance.
        """
        num_params = self.get_num_params(db)

        size = get_directory_size(self.model_save_path)
        size_in_memory = self.get_size_in_memory()
        latency = self.get_latency(db)

        self.reporter.report_complete(
            model_id=self.general_variables.model_id,
            metadata={
                "num_params": str(num_params),
                "size": str(size),
                "size_in_memory": str(size_in_memory),
                "thirdai_version": str(thirdai.__version__),
                "training_time": str(train_time),
                "latency": str(latency),
            },
        )
