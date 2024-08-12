from abc import abstractmethod
from typing import Optional

from models.model import Model
from pydantic_models import inputs
from thirdai import bolt
import logging_loki
import logging
logging_loki.emitter.LokiEmitter.level_tag = "level"
from collections import Counter

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

        self.loki_handler = logging_loki.LokiHandler(
            url="http://localhost:80/loki/api/v1/push",
            version="1",
            )

        self.db_logger = logging.getLogger("action-logger")
        self.db_logger.addHandler(self.loki_handler)
        self.db_logger.setLevel(logging.DEBUG)


    def get_udt_path(self, model_id: Optional[str] = None) -> str:
        model_id = model_id or self.general_variables.model_id
        return str(self.get_model_dir(model_id) / "model.udt")

    def load_model(self):
        return bolt.UniversalDeepTransformer.load(self.model_path)

    @abstractmethod
    def predict(self, **kwargs):
        pass


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
            deployment_id=self.general_variables.deployment_id,
            access_token=kwargs.get("token"),
            train_samples=[
                {
                    "query": query,
                    "top_k": str(top_k),
                    "class_names": ",".join(class_names),
                }
            ],
        )

        _counter = Counter(class_names)

        self.db_logger.info(
            str(self.general_variables.deployment_id),
            extra={"tags": {
                    "deployment_id": str(self.general_variables.deployment_id),
                    "service":"text_classification", 
                    "type": "udt",
                    "action": "predict",
                    **_counter,
                }}
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

    def predict(self, **kwargs):
        query = kwargs["query"]

        predicted_tags = self.model.predict({"source": query}, top_k=1)
        predictions = []
        for predicted_tag in predicted_tags:
            predictions.append([x[0] for x in predicted_tag])
        
        predictions_flattened = []
        for x in predictions: 
            predictions_flattened.extend([label for label in x if label != "O"])
        
        _counter = Counter(predictions_flattened)
        
        self.db_logger.info(
            str(self.general_variables.deployment_id),
            extra={"tags": {
                    "deployment_id": str(self.general_variables.deployment_id),
                    "service":"token_classification", 
                    "type": "udt",
                    "action": "predict",
                     **_counter,
                }}
        )
        
        return inputs.SearchResultsTokenClassification(
            query_text=query,
            tokens=query.split(),
            predicted_tags=predictions,
        )
