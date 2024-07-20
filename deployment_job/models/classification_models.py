import logging
import shutil
import tempfile
import traceback
import uuid
from abc import abstractmethod
from pathlib import Path

import numpy as np
from models.model import Model
from pydantic_models import inputs
from thirdai import bolt


class ClassificationModel(Model):
    def __init__(self):
        super().__init__()
        self.model_path = self.model_dir / "model.udt"
        self.model: bolt.UniversalDeepTransformer = self.load_model()

    def get_udt_path(self, model_id):
        return self.get_model_dir(model_id) / "model.udt"

    def load_model(self):
        return bolt.UniversalDeepTranformer.load(self.model_path)

    @abstractmethod
    def predict(self, **kwargs):
        pass


class TextClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()

    def predict(self, **kwargs):
        query = kwargs["query"]
        top_k = kwargs["top_k"]
        prediction = self.model.predict({"text": query}, top_k=top_k)
        class_names = [self.model.class_name(x) for x in prediction[0]]

        return inputs.SearchResultsTextClassification(
            query_text=query,
            class_name=class_names,
        )


class TokenClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()

    def predict(self, **kwargs):
        query = kwargs["query"]

        predicted_tags = self.model.predict({"source": query}, top_k=1)
        predicted_tags = [x[0][0] for x in predicted_tags]

        return inputs.SearchResultsTokenClassification(
            query_text=query,
            predicted_tags=predicted_tags,
        )
