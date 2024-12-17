# Workflow Management in Model Bazaar

## Create Enterprise Search Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/workflow/enterprise-search` | Yes | Read Access for all Component Models |

Creates a new Enterprise Search model.

Example Request: 

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
Example Response:
```json
{
  "model_id": "model uuid"
}
```