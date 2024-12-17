# User Management in Model Bazaar

## Signup

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/signup` | No | None |

Creates a new user. Not supported when using Keycloak authentication. Returns the user id for the new user.

Example Request: 
```json
{
  "username": "xyz",
  "email": "xyz@mail.com",
  "password": "super secret password"
}
```
Example Response:
```json
{
  "user_id": "user uuid"
}
```

## Email Login

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/login` | No | None |

Logs in as the specified user. Not supported when using Keycloak authentication. Returns the access token and user id.

Example Request: 
```json
{
  "email": "xyz@mail.com",
  "password": "password"
}
```
Example Response:
```json
{
  "user_id": "user uuid",
  "access_token": "a jwt"
}
```

## Token Login (Keycloak only)

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/login-with-token` | No | None |

Performs "login" for the user with the given access token. This seems a little counterintuitive, but this endpoint is only available when using Keycloak authentication, and it allows us to add the user to our internal DB if it doesn't already exist.

Example Request: 
```json
{
  "access_token": "<token from keycloak>"
}
```
Example Response:
```json
{
  "user_id": "user uuid",
  "access_token": "a jwt"
}
```



## List Users

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/user/list` | Yes | None |

Lists the users that are visible to the current user. If the user is an admin this is all users. If the user is not an admin this is any users that share a team with the current user.

Example Request: 
```json
```
Example Response:
Notes:
* `admin` indicatates whether or not the user is an admin. 
* `team_admin` whether or not the user is an admin for the team
```json
[
  {
    "id": "user uuid",
    "username": "xyz",
    "email": "xyz@mail.com",
    "admin": false,
    "teams": [
      {
        "team_id": "team uuid",
        "team_name": "purple team",
        "team_admin": false
      },
    ]
  },
]
```


## Get User Info

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/user/info` | yes | None |

Returns the info for the current user.

Example Request: 
```json
```
Example Response:
```json
{
  "id": "user uuid",
  "username": "xyz",
  "email": "xyz@mail.com",
  "admin": false, // or true
  "teams": [
    {
      "team_id": "team uuid",
      "team_name": "purple team",
      "team_admin": false // whether or not the user is an admin for the team
    },
    ...
  ]
}
```


## Create Users

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/create` | yes | Admin Only |

Creates a new user with the given username/email/password.

Example Request: 
```json
{
  "username": "xyz",
  "email": "xyz@mail.com",
  "password": "super secret password"
}
```
Example Response:
```json
{
  "user_id": "user uuid"
}
```

### Deleting Users

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/user/{user_id}` | Yes | Admin Only |

Deletes the specified user. Returns 200 on success. No request or response body.

Example Request: 
```json
```
Example Response:
```json
{}
```

## Promoting User to Admin

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/{user_id}/admin` | yes | Admin Only 

Promotes the specified use rto admin. Returns 200 on success. No request or response body.

Example Request: 
```json
```
Example Response:
```json
{}
```

## Demoting User from Admin

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/user/{user_id}/admin` | yes | Admin Only 

Promotes the specified use rto admin. Returns 200 on success. No request or response body.

Example Request: 
```json
```
Example Response:
```json
{}
```

## Verifying a User

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/{user_id}/verify` | yes | Admin Only 

Marks the given user as verified. Returns 200 on success, no request or response body.

Example Request: 
```json
```
Example Response:
```json
{}
```

