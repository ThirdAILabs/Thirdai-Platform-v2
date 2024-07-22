from abc import abstractmethod
from pathlib import Path

import thirdai
from models.model import Model
from thirdai import bolt
from utils import list_files
from variables import (
    TextClassificationVariables,
    TokenClassificationVariables,
    UDTVariables,
)


class ClassificationModel(Model):
    def __init__(self):
        super().__init__()
        self.udt_variables = UDTVariables.load_from_env()
        self.model_save_path = self.model_dir / "model.udt"

    @abstractmethod
    def initialize_model(self):
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

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        test_files = list_files(self.data_dir / "test")
        
        for train_file in unsupervised_files:
            model.train(
                train_file,
                epochs=self.train_variables.unsupervised_epochs,
                learning_rate=self.train_variables.learning_rate,
                batch_size=self.train_variables.batch_size,
                metrics=self.train_variables.metrics,
            )

        self.save_model(model)
        
        self.evaluate(model, test_files)
            
        self.reporter.report_complete(
            self.general_variables.model_id,
            metadata={
                "thirdai_version": str(thirdai.__version__),
            },
        )
        

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

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        test_files = list_files(self.data_dir / "test")
        
        for train_file in unsupervised_files:
            model.train(
                train_file,
                epochs=self.train_variables.unsupervised_epochs,
                learning_rate=self.train_variables.learning_rate,
                batch_size=self.train_variables.batch_size,
                metrics=self.train_variables.metrics,
            )

        self.save_model(model)

        self.evaluate(model, test_files)
        
        self.reporter.report_complete(
            self.general_variables.model_id,
            metadata={
                "thirdai_version": str(thirdai.__version__),
            },
        )