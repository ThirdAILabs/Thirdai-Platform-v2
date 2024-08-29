from abc import abstractmethod
from typing import Optional

from models.model import Model
from pydantic_models import inputs
from thirdai import bolt


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
    
    @abstractmethod
    def get_classes(self, **kwargs):
        pass


class TextClassificationModel(ClassificationModel):
    def __init__(
        self, model_id: Optional[str] = None, model_path: Optional[str] = None
    ):
        super().__init__(model_id, model_path)
        # TODO: This is hacked. Change as to use log storage
        self.num_classes = self.model.predict({"text": "test"}).shape[-1]
        self.classes = [self.model.class_name(i) for i in range(self.num_classes)]

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

    def get_classes(self, **kwargs):
        return self.classes


class TokenClassificationModel(ClassificationModel):
    def __init__(
        self, model_id: Optional[str] = None, model_path: Optional[str] = None
    ):
        super().__init__(model_id, model_path)
        # TODO: This is hacked. Change as to use log storage
        top_k = ((self.model._get_model().num_params() - 2000*100000) - 2000) // 2001 - 1
        print(self.model.predict({"source": "hello"}, top_k=top_k)[0])
        self.classes = list(set(tag for tag, score in self.model.predict({"source": "hello"}, top_k=top_k)[0] if tag != 'O'))

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
    
    def get_classes(self, **kwargs):
        return self.classes
