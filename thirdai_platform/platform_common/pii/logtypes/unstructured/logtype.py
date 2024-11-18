from platform_common.pii.logtypes.pydantic_models import (
    UnstructuredTokenClassificationResults,
)


class UnstructuredTokenClassificationLog:
    def __init__(self, log: str):
        self.inference_sample = {"source": log}

    def process_prediction(self, model_predictions: str):
        predictions = []
        for prediction in model_predictions:
            predictions.append([x[0] for x in prediction])

        tokens = self.inference_sample["source"].split()

        return UnstructuredTokenClassificationResults(
            literal="unstructured",
            query_text=self.inference_sample["source"],
            tokens=tokens,
            predicted_tags=predictions,
        )
