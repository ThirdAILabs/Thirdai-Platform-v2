# Model Management in Model Bazaar

## Get Model Info 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/model/{model_id}` | Yes | Model Read Access Only |

Gets info about the specified model.

__Example Request__: 
```json
```
__Example Response__:

Notes:
* `attributes` and `dependencies` may be empty. 
* `team_id` will be null if the model is not assigned to a team.
```json
{
  "model_id": "model uuid",
  "model_name": "my-model",
  "type": "enterprise-search",
  "access": "private",
  "train_status": "complete",
  "deploy_status": "in_progress",
  "publish_date": "2014-05-16T08:28:06.801064-04:00",
  "user_email": "my-model-owner@mail.com",
  "username": "my-model-owner",
  "team_id": "model team uuid",
  "attributes": {
    "key": "value"
  }, 
  "dependencies": [
    {
      "model_id": "dependency model uuid",
      "model_name": "my-dependent-model",
      "type": "ndb",
      "username": "my-model-owner"
    }
  ]
}
```

## Delete a Model

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/model/{model_id}` | Yes | Model Owner Only |

Deletes the specified model. No request or response body. Returns 200 on success.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## Update Model Access Level 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/model/{model_id}/access` | Yes | Model Owner Only |

Updates the access level of the given model. If the new access level is `'protected'` then a second request parameter `'team_id'` must be specified to indicate which team the model should added to. If the model is already part of a different team, this will change which team the model belongs to. If the user is not an system admin then they must be a member of the team that the model is being added to.

__Example Request__: 
```json
{
  "access": "protected",
  "team_id": "team uuid"
}
```
__Example Response__:
```json
{}
```

## Update Model Default Permission

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/model/{model_id}/default-permission` | Yes | Model Owner Only |

Updates the default permission of the given model.

__Example Request__: 
```json
{
  "permission": "read"
}
```
__Example Response__:
```json
{}
```

## List Models 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/model/list` | Yes | None |

Returns a list of models that are accessible to the current user. If the user is an admin this is all models, otherwise it is the models a user owns, or the protected models that are assigned to one of the user's teams.

__Example Request__: 
```json
```
__Example Response__:

Notes:
* `attributes` and `dependencies` may be empty. 
* `team_id` will be null if the model is not assigned to a team.
```json
[
  {
    "model_id": "model uuid",
    "model_name": "my-model",
    "type": "enterprise-search",
    "access": "private",
    "train_status": "complete",
    "deploy_status": "in_progress",
    "publish_date": "2014-05-16T08:28:06.801064-04:00",
    "user_email": "my-model-owner@mail.com",
    "username": "my-model-owner",
    "team_id": "model team uuid",
    "attributes": {
      "key": "value"
    }, 
    "dependencies": [
      {
        "model_id": "dependency model uuid",
        "model_name": "my-dependent-model",
        "type": "ndb",
        "username": "my-model-owner"
      }
    ]
  }
]
```

## Get Model Permissions for a User 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/model/{model_id}/permissions` | Yes | None |

Returns the permissions of the current user for the given model.

__Example Request__: 
```json
```
__Example Response__:
```json
{
  "read": true,
  "write": true,
  "owner": false,
  "exp": "2014-05-16T08:28:06.801064-04:00"
}
```

## Download a Model 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/model/{model_id}/download` | Yes | Model Read Access Only |

Returns the raw model. Sends the data in chunks.

__Example Request__: 
```json
```
__Example Response__:
```
Raw model data.
```

## Begin Model Upload

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/model/upload` | Yes | None |

Creates an entry for a new model that will be uploaded. Returns an upload session token that must be used to upload chunks and complete the upload.

__Example Request__: 
```json
{
  "model_name": "name of new model",
  "model_type": "ndb"
}
```
__Example Response__:
```json
{
  "token": "upload token"
}
```

## Upload Model Chunk 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/model/upload/{chunk_idx}` | Yes (Upload Session Token) | Must have Upload Session Token |

Stores the given chunk data as part of the upload of the new model. The url parameter `chunk_idx` is used to order the chunks. Returns 200 on success.

__Example Request__: 
```
Raw data of chunk
```
__Example Response__:
```json
{}
```

## Complete Model Upload 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/model/upload/commit` | Yes (Upload Session Token) | Must have Upload Session Token |

Completes the model upload and updates the model train status to complete to indicate the model can be used. Returns the uuid of the new model.

__Example Request__: 
```json
```
__Example Response__:
```json
{
  "model_id": "new model uuid"
}
```