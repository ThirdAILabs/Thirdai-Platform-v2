from abc import abstractmethod

import numpy as np
from models.model import Model
from pydantic_models import inputs
from thirdai import bolt


class ClassificationModel(Model):
    def __init__(self):
        super().__init__()
        self.model_path = self.model_dir / "model.udt"
        self.model = bolt.UniversalDeepTransformer = self.load_model()
        
    def load_model(self):
        return bolt.UniversalDeepTranformer.load(self.model_path)
    
    @abstractmethod
    def predict(self, **kwargs):
        pass
    
class TextClassificationModel(ClassificationModel):
    def __init__(self):
        super().__init__()
    
    def predict(self, **kwargs):
        query = kwargs['query']
        
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
        query = kwargs['query']
        
        predicted_tags = self.model.predict({"source": query}, top_k=1)
        predicted_tags = [x[0][0] for x in predicted_tags]
        
        return inputs.SearchResultsTokenClassification(
            query_text=query,
            predicted_tags=predicted_tags,
        )
                