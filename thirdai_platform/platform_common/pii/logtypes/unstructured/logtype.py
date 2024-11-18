from platform_common.pii.logtypes.base import LogType
from platform_common.pii.logtypes.pydantic_models import (
    UnstructuredTokenClassificationResults,
)


class UnstructuredTokenClassificationLog(LogType):
    def __init__(self, log: str):
        self.inference_sample = {"source": log}

    def process_prediction(self, model_predictions: str):
        predictions = []
        for prediction in model_predictions:
            predictions.append([x[0] for x in prediction])

        tokens = self.inference_sample["source"].split()

        if len(predictions) != len(tokens):
            raise ValueError(
                "Error parsing input text, this is likely because the input contains unsupported unicode characters."
            )

        return UnstructuredTokenClassificationResults(
            literal="unstructured",
            query_text=self.inference_sample["source"],
            tokens=tokens,
            predicted_tags=predictions,
        )
