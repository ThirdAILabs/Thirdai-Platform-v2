# Deploying Models in Model Bazaar

## Deploy a Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/deploy/{model_id}` | Yes | Model Owner Only |

Deploys the specified model. Returns 200 on success.

__Example Request__: 

Notes:
* All parameters are optional.
* `deployment_name` is used to set a custom url for the deployment. 
```json
{
  "deployment_name": "my-app",
  "autoscaling_enabled": true,
  "autoscaling_min": 1,
  "autoscaling_max": 4,
  "memory": 800
}
```
__Example Response__:
```json
{}
```

## Undeploy a Model

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/deploy/{model_id}` | Yes | Model Owner Only |

Undeploys the specified model. Returns 200 on success. No request or response body.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## Get Deployment Status

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/deploy/{model_id}/status` | Yes | Model Read Access Only |

Returns the deploy status of the model.

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

## Get Deployment Logs

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/deploy/{model_id}/logs` | Yes | Model Read Access Only |

Returns the recent logs from the model deployment.

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

# Internal Only Methods

## Save Deployed Model

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/deploy/{model_id}/save` | Yes | Model Read Access Only |

Creates a new model entry for save a deployed model to. The new model's uuid is returned along with a update token. The new model's train status is set to `in_progress` so that it cannot be deployed. The update token can by used to then update the status to `complete` after the save is done. This should only be called by the deployment job.

__Example Request__: 
```json
{
  "model_name": "name for new saved model"
}
```
__Example Response__:
```json
{
  "model_id": "new model uuid",
  "update_token": "job token"
}
```

## Get Deployment Status From Deploy Job

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/deploy/{model_id}/status-internal` | Yes (Job Auth) | Job Auth Token Required |

Returns the deploy status of the model.

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

## Update Deployment Status

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/deploy/update-status` | Yes (Job Auth) | Job Auth Token Required |

Updates the deploy status of the model. The model is determined by looking at the model associated with the job token. If specified the metadata set as a json string in the `metadata` attribute of the model. This should only be called by the deployment job.

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


## Log Deployment Error/Warning

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/deploy/log` | Yes (Job Auth) | Job Auth Token Required |

Logs a warning/error message for the deployment job. The model is determined by looking at the model associated with the job token. Returns 200 on success. This should only be called by the deployment job.

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