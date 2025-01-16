# Training Models in Model Bazaar

## Train NDB Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/ndb` | Yes | Read Access For Base Model (if specified) |

Trains a NDB model.

__Example Request__: 

Notes:
* `base_model_id` is optional, indicates a base model to start from.
* `model_options` are optional. Defaults will be used if not specified. Cannot be specified if `base_model_id` is specified.
* Both `unsupervised_files` and `supervised_files` cannot be empty in `data` field, but other args are optional.
* All fields within `job_options` are optional and have defaults.
```json
{
  "model_name": "my-model",
  "base_model_id": null,
  "model_options": {
    "in_memory": false,
    "advanced_search": false
  },
  "data": {
    "unsupervised_files": [
      {
        "path": "/path/to/file.pdf",
        "location": "s3",
      }
    ],
    "supervised_files": [],
    "deletions": ["doc_id_1"]
  },
  "job_options": {
    "allocation_cores": 4,
    "allocation_memory": 2000
  }
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Retrain NDB Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/ndb-retrain` | Yes | Read Access For Base Model |

Retrains the NDB model from all feedback collected from users.

__Example Request__: 

Notes:
* `base_model_id` is required in this endpoint.
* All fields within `job_options` are optional and have defaults.
```json
{
  "model_name": "my-model",
  "base_model_id": "my-base-model",
  "job_options": {
    "allocation_cores": 4,
    "allocation_memory": 2000
  }
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Train NLP Token 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/nlp-token` | Yes | Read Access For Base Model (if specified) |

Trains a NLP token model on the given data.

__Example Request__: 

Notes:
* `base_model_id` is optional, indicates a base model to start from.
* In `model_options`, `source_column` and `target_column` must be specfied. Other args are optional and have defaults.
* `test_files` is optional in data fields.
* All fields within `train_options` are optional and have defaults.
* `job_options` is optional.
```json
{
  "model_name": "my-model",
  "base_model_id": null,
  "model_options": {
    "target_labels": ["PHONE", "EMAIL"],
    "source_column": "source",
    "target_column": "target",
    "default_tag": "O"
  },
  "data": {
    "supervised_files": [
      {
        "path": "/path/to/train.csv",
        "location": "s3",
      }
    ],
    "test_files": [
      {
        "path": "/path/to/test.csv",
        "location": "s3"
      }
    ]
  },
  "train_options": {
    "epochs": 5,
    "learning_rate": 0.0001,
    "batch_size": 1000,
    "max_in_memory_batches": 20,
    "test_split": 0.2
  },
  "job_options": {
    "allocation_cores": 4,
    "allocation_memory": 2000
  }
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Train NLP Text 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/nlp-text` | Yes | Read Access For Base Model (if specified) |

__Example Request__: 

Notes:
* `base_model_id` is optional, indicates a base model to start from.
* Arg `doc_classification` defaults to false.
* In `model_options`, `text_column`, `label_column`, and `n_target_classes` must be specfied. Other args are optional and have defaults. Unless `doc_classification` is true, then only `n_target_classes` is required.
* `test_files` is optional in data fields.
* All fields within `train_options` are optional and have defaults.
* All fields within `job_options` are optional and have defaults.
```json
{
  "model_name": "my-model",
  "doc_classification": false,
  "base_model_id": null,
  "model_options": {
    "text_column": "source",
    "label_column": "target",
    "n_target_classes": 10,
    "delimiter": ","
  },
  "data": {
    "supervised_files": [
      {
        "path": "/path/to/train.csv",
        "location": "local",
      }
    ],
    "test_files": [
      {
        "path": "/path/to/test.csv",
        "location": "local"
      }
    ]
  },
  "train_options": {
    "epochs": 5,
    "learning_rate": 0.0001,
    "batch_size": 1000,
    "max_in_memory_batches": 20,
    "test_split": 0.2
  },
  "job_options": {
    "allocation_cores": 4,
    "allocation_memory": 2000
  }
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Train NLP with Datagen 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/nlp-datagen` | Yes | Read Access For Base Model (if specified) |

__Example Request__: 

Notes:
* `base_model_id` is optional, indicates a base model to start from.
* Exactly one of `token_options` and `text_options` must be specified.
* All fields within `train_options` are optional and have defaults.
* In `token_options`: 
  * Only the `tags` field is required. 
  * Only the `name` field must be specified for each tag. 
* In `text_options`: 
  * Only the `labels` and `samples-per_label` fields are required. 
  * Only the `name` field must be specified for each label. 
* All fields within `train_options` are optional and have defaults.
* All fields within `job_options` are optional and have defaults.
```json
{
  "model_name": "my-model",
  "base_model_id": null,
  "task_prompt": "description of task",
  "llm_provider": "openai",
  "test_size": 0.1,
  "token_options": {
    "tags": [
      {
        "name": "PHONE",
        "examples": [
          "my phone number is 123-456-7890"
        ],
        "description": "a phone number",
        "status": "uninserted"
      }
    ],
    "num_sentences_to_generate": 2000,
    "num_samples_per_tag": 10,
    "samples": [
      {
        "tokens": ["my", "number", "is", "123-456-7890"],
        "tags": ["O", "O", "O", "PHONE"]
      }
    ],
    "templates_per_sample": 4
  },
  "text_options": {
    "labels": [
      {
        "name": "positive",
        "examples": [
          "i love cookies"
        ],
        "description": "a positive statement",
        "status": "uninserted"
      }
    ],
    "samples_per_label": 10,
  },
  "train_options": {
    "epochs": 5,
    "learning_rate": 0.0001,
    "batch_size": 1000,
    "max_in_memory_batches": 20,
    "test_split": 0.2
  },
  "job_options": {
    "allocation_cores": 4,
    "allocation_memory": 2000
  }
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Retrain NLP Token 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/nlp-token-retrain` | Yes | Read Access For Base Mode |

Retrains the given NLP token model based off of collected user feedback.

__Example Request__: 

Notes:
* `base_model_id` is required in this endpoint.
* `llm_provider` and `test_size` are optional and have defaults.
* All fields within `train_options` are optional and have defaults.
* All fields within `job_options` are optional and have defaults.
```json
{
  "model_name": "my-model",
  "base_model_id": "my-base-model",
  "llm_provider": "openai",
  "test_size": 0.1,
  "train_options": {
    "epochs": 5,
    "learning_rate": 0.0001,
    "batch_size": 1000,
    "max_in_memory_batches": 20,
    "test_split": 0.2
  },
  "job_options": {
    "allocation_cores": 4,
    "allocation_memory": 2000
  }
}
```
__Example Response__:
```json
{
  "model_id": "model uuid"
}
```

## Upload Data 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/upload-data` | Yes | None |

Accepts a multipart request containing multiple files to upload for training. Returns an upload uuid that can be provided to a train endpoint to train on the files. For example if the upload id is `abc` then a user could pass `{"location": "upload", "path": "abc"}` to train on the file(s) in the upload. Note that only the user who created the upload can use that upload in training.

__Example Request__: 
```
A multipart request containing the files.
```
__Example Response__:
```json
{
  "upload_id": "uuid"
}
```

## Get Train Status

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/train/{model_id}/status` | Yes | Model Read Access Only |

Returns the train status of the model.

__Example Request__: 
```json
```
__Example Response__:
```json
{
  "status": "complete",
  "errors": [],
  "warnings": [
    "a warning message"
  ]
}
```

## Get Train Logs

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/train/{model_id}/logs` | Yes | Model Read Access Only |

Returns the recent logs from the model training.

__Example Request__: 
```json
```
__Example Response__:

Notes: 
* The output is a list because it will return a set of logs for each allocation.
```json
[
  {
    "stdout": "logs from stdout", 
    "stderr": "logs from stderr"
  }
]
```

## Get Train Report 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/train/{model_id}/report` | Yes | Model Read Access Only |

Returns the train report for the most recent training. This is currently only supported for certain types of models that create a report.

__Example Request__: 
```json
```
__Example Response__:
```
Train report json
```

# Internal Only Methods

## Update Train Status

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/train/update-status` | Yes (Job Auth) | Job Auth Token Required |

Updates the train status of the model. The model is determined by looking at the model associated with the job token. If specified the metadata set as a json string in the `metadata` attribute of the model. This should only be called by the train job.

__Example Request__: 

Notes: 
* `metadata` is optional.
```json
{
  "status": "in_progress",
  "metadata": {
    "key": "value"
  }
}
```
__Example Response__:
```json
{}
```


## Log Train Error/Warning

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/train/log` | Yes (Job Auth) | Job Auth Token Required |

Logs a warning/error message for the train job. The model is determined by looking at the model associated with the job token. Returns 200 on success. This should only be called by the train job.

__Example Request__: 
```json
{
  "level": "error", 
  "message": "this failed"
}
```
__Example Response__:
```json
{}
```