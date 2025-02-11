# Team Management in Model Bazaar

## Create Team

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/team/create` | Yes | Admin Only |

Creates a new team.

__Example Request__: 
```json
{
  "name": "name of team"
}
```
__Example Response__:
```json
{
  "team_id": "team uuid"
}
```

## Delete Team

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/team/{team_id}` | Yes | Admin Only |

Deletes a team.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## List Teams 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/team/list` | Yes | None |

Lists all teams visible to the current user. If the use is an admin then this is all teams. Otherwise it is the teams the user is a member of.

__Example Request__: 
```json
```
__Example Response__:
```json
[
  {
    "id": "team uuid",
    "name": "team name"
  }
]
```

## List Team Users

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/team/{team_id}/users` | Yes | Team Admin Only |

Lists the users in the team.

__Example Request__: 
```json
```
__Example Response__:

Notes:
* `team_admin` indicates if the user is an admin of the team.
```json
[
  {
    "user_id": "user uuid",
    "username": "xyz", 
    "email": "xyz@mail.com",
    "team_admin": false
  }
]
```

## List Team Models

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/team/{team_id}/models` | Yes | Team Admin Only |

Lists the models in the team.

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

## Add a User to a Team 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/team/{team_id}/users/{user_id}` | Yes | Team Admin Only |

Adds the user as a member of the team.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## Remove a User from a Team 

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/team/{team_id}/users/{user_id}` | Yes | Team Admin Only |

Removes the user from the team.

__Example Request__: 
```json

```
__Example Response__:
```json
{}
```

## Add a Team Admin

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/team/{team_id}/admins/{user_id}` | Yes | Team Admin Only |

Adds a user as an admin for the team.

__Example Request__: 
```json

```
__Example Response__:
```json
{}
```

## Remove a Team Admin

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/team/{team_id}/admins/{user_id}` | Yes | Team Admin Only |

Removes an admin from the team. The user will still be a member of the team.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```