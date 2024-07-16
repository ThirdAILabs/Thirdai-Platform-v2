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
        model.save(self.model_save_path)

    def get_model(self):
        if self.udt_variables.base_model_id:
            return self.load_model(self.udt_variables.base_model_id)
        return self.initialize_model()

    def train(self, **kwargs):
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        model = self.get_model()

        unsupervised_files = list_files(self.data_dir / "unsupervised")

        if unsupervised_files:
            for train_file in unsupervised_files:
                model.train(
                    train_file,
                    epochs=3,
                    learning_rate=0.001,
                    batch_size=1024,
                    metrics=["loss", "categorical_accuracy"],
                )

        self.save_model(model)

        self.reporter.report_complete(
            self.general_variables.model_id,
            metadata={
                "thirdai_version": str(thirdai.__version__),
            },
        )


class TextClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()
        self.classification_vars = TextClassificationVariables.load_from_env()

    def initialize_model(self):
        return bolt.UniversalDeepTransformer(
            data_types={
                "text": bolt.types.text(),
                "label": bolt.types.categorical(
                    n_classes=self.text_classification_vars.n_target_classes
                ),
            },
            target="label",
            delimiter=self.text_classification_vars.delimiter,
        )

    def evaluate(self, **kwargs):
        pass


class TokenClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()
        self.classification_vars = TextClassificationVariables.load_from_env()

    def initialize_model(self):
        target_labels = self.token_classification_vars.target_labels
        default_tag = self.token_classification_vars.default_tag
        return bolt.UniversalDeepTransformer(
            data_types={
                "source": bolt.types.text(),
                "target": bolt.types.token_tags(
                    tags=target_labels, default_tag=default_tag
                ),
            },
            target="target",
        )

    def evaluate(self, **kwargs):
        pass
