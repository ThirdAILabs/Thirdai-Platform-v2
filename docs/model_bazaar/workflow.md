# Workflow Management in Model Bazaar

## Create Enterprise Search Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/workflow/enterprise-search` | Yes | Read Access for all Component Models |

Creates a new Enterprise Search model.

__Example Request__: 

Notes:
* All args are optional except for `model_name` and `retrieval_id`.
```json
{
  "model_name": "my-search",
  "retrieval_id": "uuid for ndb component",
  "guardrail_id": "uuid for guardrail component",
  "llm_provider": "openai",
  "nlp_classifier_id": "uuid for classifier component",
  "default_mode": "chat"
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Create Knowledge Extraction Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/workflow/knowlege-extraction` | Yes | Read Access for all Component Models |

Creates a new Knowledge Extraction model.

__Example Request__: 

Notes:
* All args are required except for `advanced_indexing`, `rerank`, and `generate_answers`.
* At least 1 question must be specified, though question keywords are optional.
* Error will be returned on duplicate questions. Duplicate is defined here as case insensitive exact match. 
```json
{
  "model_name": "my-search",
  "questions": [
    {
      "question": "what are the earnings per share",
      "keywords": ["EPS"],
    }
  ],
  "llm_provider": "openai",
  "advanced_indexing": true,
  "rerank": true,
  "generate_answers": true,
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```