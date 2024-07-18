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

    def save_udt(self, **kwargs):
        model_path = self.get_udt_path(kwargs.get("model_id"))
        temp_dir = None

        try:
            with tempfile.TemporaryDirectory() as temp_dir:
                temp_model_path = Path(temp_dir) / "model.udt"
                self.model.save(save_to=temp_model_path)
                if model_path.exists():
                    backup_id = str(uuid.uuid4())
                    backup_path = self.get_udt_path(backup_id)
                    print(f"Creating backup: {backup_id}")
                    shutil.copytree(model_path, backup_path)

                if model_path.exists():
                    shutil.rmtree(model_path)
                shutil.move(temp_model_path, model_path)

                if model_path.exists() and "backup_path" in locals():
                    shutil.rmtree(backup_path.parent)

        except Exception as err:
            logging.error(f"Failed while saving with error: {err}")
            traceback.print_exc()

            if "backup_path" in locals() and backup_path.exists():
                if model_path.exists():
                    shutil.rmtree(model_path)
                shutil.copytree(backup_path, model_path)
                shutil.rmtree(backup_path.parent)

            raise

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

        prediction = self.model.predict({"text": query})
        class_name = self.model.class_name(np.argmax(prediction))

        return inputs.SearchResultsTextClassification(
            query_text=query,
            class_name=class_name,
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
