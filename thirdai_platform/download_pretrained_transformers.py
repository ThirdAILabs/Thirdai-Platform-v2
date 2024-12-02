from transformers import AutoModelForMaskedLM, AutoModelForSequenceClassification

# The following models are downloaded in Universe. If we change the models we use in Universe,
# we must also change the models that we load here.

model = AutoModelForSequenceClassification.from_pretrained(
    "jinaai/jina-reranker-v1-tiny-en", num_labels=1, trust_remote_code=True
)

model = AutoModelForMaskedLM.from_pretrained("naver/splade-cocondenser-selfdistil")
