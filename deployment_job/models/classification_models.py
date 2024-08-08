from abc import abstractmethod

from models.model import Model
from pydantic_models import inputs
from thirdai import bolt
import logging_loki
import logging
logging_loki.emitter.LokiEmitter.level_tag = "level"
from collections import Counter

class ClassificationModel(Model):
    def __init__(self):
        super().__init__()
        self.model_path = self.get_udt_path()
        self.model: bolt.UniversalDeepTransformer = self.load_model()

        self.loki_handler = logging_loki.LokiHandler(
            url="http://192.168.1.11/loki/api/v1/push",
            version="1",
            )

        self.db_logger = logging.getLogger("action-logger")
        self.db_logger.addHandler(self.loki_handler)
        self.db_logger.setLevel(logging.DEBUG)


    def get_udt_path(self):
        return str(self.get_model_dir(self.general_variables.model_id) / "model.udt")

    def load_model(self):
        return bolt.UniversalDeepTransformer.load(self.model_path)

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
    def __init__(self):
        super().__init__()

    def predict(self, **kwargs):
        query = kwargs["query"]
        top_k = kwargs["top_k"]

        predicted_tags = self.model.predict({"source": query}, top_k=top_k)
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
            predicted_tags=predictions,
        )
