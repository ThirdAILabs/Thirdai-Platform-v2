from transformers import AutoModelForMaskedLM, AutoModelForSequenceClassification

# Download Splade model
print("Downloading Splade model...")
AutoModelForMaskedLM.from_pretrained("naver/splade-cocondenser-selfdistil")
print("Splade model downloaded.")

# Download Jina Reranker model
print("Downloading Jina Reranker model...")
AutoModelForSequenceClassification.from_pretrained(
    "jinaai/jina-reranker-v1-tiny-en", num_labels=1, trust_remote_code=True
)
print("Jina Reranker model downloaded.")
