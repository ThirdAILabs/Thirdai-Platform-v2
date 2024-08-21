import time
from abc import abstractmethod
from pathlib import Path

import thirdai
from exceptional_handler import apply_exception_handler
from models.model import Model
from thirdai import bolt
from utils import list_files
from variables import TextClassificationVariables, TokenClassificationVariables


@apply_exception_handler
class ClassificationModel(Model):
    report_failure_method = "report_status"

    def __init__(self):
        super().__init__()
        self.model_save_path = self.model_dir / "model.udt"

    @abstractmethod
    def initialize_model(self):
        pass

    def get_size(self):
        """
        Calculate the size of the model in bytes
        """
        return self.model_save_path.stat().st_size

    def get_num_params(self, model: bolt.UniversalDeepTransformer) -> int:
        """
        Gets the number of parameters in the model

        Args:
            model: (bolt.UniversalDeepTransformer): The UDT instance
        Returns:
            int: The number of parameters in the model.

        """

        num_params = model._get_model().num_params()
        self.logger.info(f"Number of parameters in the model: {num_params}")
        return num_params

    @abstractmethod
    def get_latency(self):
        pass

    def get_udt_path(self, model_id):
        return (
            Path(self.general_variables.model_bazaar_dir)
            / "models"
            / model_id
            / "model.udt"
        )

    def load_model(self, model_id):
        return bolt.UniversalDeepTransformer.load(self.get_udt_path(model_id))

    def save_model(self, model):
        model.save(str(self.model_save_path))

    def get_model(self):
        if self.general_variables.base_model_id:
            return self.load_model(self.general_variables.base_model_id)
        return self.initialize_model()

    def evaluate(self, model, test_files):
        for test_file in test_files:
            model.evaluate(test_file, metrics=self.train_variables.validation_metrics)

    @abstractmethod
    def train(self, **kwargs):
        pass


@apply_exception_handler
class TextClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()
        self.txt_cls_vars = TextClassificationVariables.load_from_env()

    def initialize_model(self):
        return bolt.UniversalDeepTransformer(
            data_types={
                self.txt_cls_vars.text_column: bolt.types.text(),
                self.txt_cls_vars.label_column: bolt.types.categorical(
                    n_classes=self.txt_cls_vars.n_target_classes
                ),
            },
            target=self.txt_cls_vars.label_column,
            delimiter=self.txt_cls_vars.delimiter,
        )

    def train(self, **kwargs):
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        model = self.get_model()

        supervised_files = list_files(self.data_dir / "supervised")
        test_files = list_files(self.data_dir / "test")

        start_time = time.time()
        for train_file in supervised_files:
            if not train_file.endswith("nfs_files.txt"):
                model.train(
                    train_file,
                    epochs=self.train_variables.unsupervised_epochs,
                    learning_rate=self.train_variables.learning_rate,
                    batch_size=self.train_variables.batch_size,
                    metrics=self.train_variables.metrics,
                )
            else:
                for line in train_file:
                    actual_train_file = line.strip()
                    model.train(
                        actual_train_file,
                        epochs=self.train_variables.unsupervised_epochs,
                        learning_rate=self.train_variables.learning_rate,
                        batch_size=self.train_variables.batch_size,
                        metrics=self.train_variables.metrics,
                    )
                    
        training_time = time.time() - start_time

        self.save_model(model)

        self.evaluate(model, test_files)

        num_params = self.get_num_params(model)
        model_size = self.get_size()
        model_size_in_memory = model_size * 4
        latency = self.get_latency(model)

        self.reporter.report_complete(
            self.general_variables.model_id,
            metadata={
                "num_params": str(num_params),
                "thirdai_version": str(thirdai.__version__),
                "training_time": str(training_time),
                "size": str(model_size),
                "size_in_memory": str(model_size_in_memory),
                "latency": str(latency),
            },
        )

    def get_latency(self, model) -> float:

        self.logger.info("Measuring latency of the UDT instance.")

        start_time = time.time()
        model.predict({self.txt_cls_vars.text_column: "Checking for latency"}, top_k=1)
        latency = time.time() - start_time

        self.logger.info(f"Latency measured: {latency} seconds.")
        return latency


@apply_exception_handler
class TokenClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()
        self.tkn_cls_vars = TokenClassificationVariables.load_from_env()

    def initialize_model(self):
        target_labels = self.tkn_cls_vars.target_labels
        default_tag = self.tkn_cls_vars.default_tag
        return bolt.UniversalDeepTransformer(
            data_types={
                self.tkn_cls_vars.source_column: bolt.types.text(),
                self.tkn_cls_vars.target_column: bolt.types.token_tags(
                    tags=target_labels, default_tag=default_tag
                ),
            },
            target=self.tkn_cls_vars.target_column,
        )

    def train(self, **kwargs):
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        model = self.get_model()

        supervised_files = list_files(self.data_dir / "supervised")
        test_files = list_files(self.data_dir / "test")

        start_time = time.time()
        for train_file in supervised_files:
            if not train_file.endswith("nfs_files.txt"):
                model.train(
                    train_file,
                    epochs=self.train_variables.unsupervised_epochs,
                    learning_rate=self.train_variables.learning_rate,
                    batch_size=self.train_variables.batch_size,
                    metrics=self.train_variables.metrics,
                )
            else:
                for line in open(train_file):
                    actual_train_file = line.strip()
                    model.train(
                        actual_train_file,
                        epochs=self.train_variables.unsupervised_epochs,
                        learning_rate=self.train_variables.learning_rate,
                        batch_size=self.train_variables.batch_size,
                        metrics=self.train_variables.metrics,
                    )
        training_time = time.time() - start_time

        self.save_model(model)

        self.evaluate(model, test_files)

        num_params = self.get_num_params(model)
        model_size = self.get_size()
        model_size_in_memory = model_size * 4
        latency = self.get_latency(model)

        self.reporter.report_complete(
            self.general_variables.model_id,
            metadata={
                "num_params": str(num_params),
                "thirdai_version": str(thirdai.__version__),
                "training_time": str(training_time),
                "size": str(model_size),
                "size_in_memory": str(model_size_in_memory),
                "latency": str(latency),
            },
        )

    def get_latency(self, model) -> float:

        self.logger.info("Measuring latency of the UDT instance.")

        start_time = time.time()
        model.predict(
            {self.tkn_cls_vars.source_column: "Checking for latency"}, top_k=1
        )
        latency = time.time() - start_time

        self.logger.info(f"Latency measured: {latency} seconds.")
        return latency
