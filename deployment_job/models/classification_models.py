from abc import abstractmethod
from typing import Optional

from models.model import Model
from pydantic_models import inputs
from thirdai import bolt

from storage.storage import DataStorage, SQLiteConnector
from storage.data_types import TokenClassificationSample, TagMetadata, DataSample


class ClassificationModel(Model):
    def __init__(
        self, model_id: Optional[str] = None, model_path: Optional[str] = None
    ):
        super().__init__()
        if model_path:
            self.model_path = model_path
        else:
            self.model_path = self.get_udt_path(model_id)
        self.model: bolt.UniversalDeepTransformer = self.load_model()

    def get_udt_path(self, model_id: Optional[str] = None) -> str:
        model_id = model_id or self.general_variables.model_id
        return str(self.get_model_dir(model_id) / "model.udt")

    def load_model(self):
        return bolt.UniversalDeepTransformer.load(self.model_path)

    @abstractmethod
    def predict(self, **kwargs):
        pass

    def get_labels(self):
        raise NotImplementedError(
            f"The method 'get_labels' is not implemented for the model class type: {self.__class__.__name__}"
        )

    def add_labels(self, label):
        raise NotImplementedError(
            f"The method 'add_labels' is not implemented for the model class type: {self.__class__.__name__}"
        )

    def insert_sample(self, sample: inputs.BaseModel):
        raise NotImplementedError(
            f"The method 'insert_sample' is not implemented for the model class type: {self.__class__.__name__}"
        )


class TextClassificationModel(ClassificationModel):
    def __init__(
        self, model_id: Optional[str] = None, model_path: Optional[str] = None
    ):
        super().__init__(model_id, model_path)

    def predict(self, **kwargs):
        query = kwargs["query"]
        top_k = kwargs["top_k"]
        prediction = self.model.predict({"text": query}, top_k=top_k)
        class_names = [self.model.class_name(x) for x in prediction[0]]

        self.reporter.log(
            action="predict",
            model_id=self.general_variables.model_id,
            access_token=kwargs.get("token"),
            train_samples=[
                {
                    "query": query,
                    "top_k": str(top_k),
                    "class_names": ",".join(class_names),
                }
            ],
        )

        return inputs.SearchResultsTextClassification(
            query_text=query,
            class_names=class_names,
        )


class TokenClassificationModel(ClassificationModel):
    def __init__(
        self, model_id: Optional[str] = None, model_path: Optional[str] = None
    ):
        super().__init__(model_id, model_path)

        self.load_storage()

    def predict(self, **kwargs):
        query = kwargs["query"]

        predicted_tags = self.model.predict({"source": query}, top_k=1)
        predictions = []
        for predicted_tag in predicted_tags:
            predictions.append([x[0] for x in predicted_tag])

        self.reporter.log(
            action="predict",
            model_id=self.general_variables.model_id,
            access_token=kwargs.get("token"),
            train_samples=[{"query": query, "predictions": ",".join(predictions[0])}],
        )

        return inputs.SearchResultsTokenClassification(
            query_text=query,
            tokens=query.split(),
            predicted_tags=predictions,
        )

    def load_storage(self):
        data_storage_path = self.data_dir / "data_storage.db"
        # connector will instantiate an sqlite db at the specified path if it doesn't exist
        self.data_storage = DataStorage(
            connector=SQLiteConnector(db_path=data_storage_path)
        )

    def get_labels(self):
        # load tags and their status from the storage
        tag_metadata = self.data_storage.get_metadata("tags_and_status")
        return list(tag_metadata._tag_and_status.keys())

    def add_labels(self, labels):
        tag_metadata = self.data_storage.get_metadata("tags_and_status")

        for label in labels:
            tag_metadata.update_tag_status(label, "untrained")

        # update the metadata entry in the DB
        self.data_storage.insert_metadata(tag_metadata)

    def insert_sample(self, sample: TokenClassificationSample):
        token_tag_sample = DataSample(name="ner", sample=sample)
        self.data_storage.insert_samples(
            samples=[token_tag_sample], override_buffer_limit=True
        )
