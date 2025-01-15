# User Management in Model Bazaar

## Signup

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/signup` | No | None |

Creates a new user. Not supported when using Keycloak authentication. Returns the user id for the new user.

__Example Request__: 
```json
{
  "username": "xyz",
  "email": "xyz@mail.com",
  "password": "super secret password"
}
```
__Example Response__:
```json
{
  "user_id": "user uuid"
}
```

## Email Login

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `GET` | `/api/v2/user/login` | No | None |

Logs in as the specified user. Not supported when using Keycloak authentication. Returns the access token and user id. The email and password should be passed in the `Authorization` header using the scheme `Basic`. See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization for details for the format of the header, though most libraries will have support for constructing the correct header from the email and password. for example `req.SetBasicAuth(email, password)` in Go.

__Example Request__: 
```json
```
__Example Response__:
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

__Example Request__: 
```json
{
  "access_token": "<token from keycloak>"
}
```
__Example Response__:
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

__Example Request__: 
```json
```
__Example Response__:

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

__Example Request__: 
```json
```
__Example Response__:

Notes:
* `admin` indicatates whether or not the user is an admin. 
* `team_admin` whether or not the user is an admin for the team
```json
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
}
```


## Create Users

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/create` | yes | Admin Only |

Creates a new user with the given username/email/password.

__Example Request__: 
```json
{
  "username": "xyz",
  "email": "xyz@mail.com",
  "password": "super secret password"
}
```
__Example Response__:
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

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## Promoting User to Admin

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/{user_id}/admin` | yes | Admin Only 

Promotes the specified use rto admin. Returns 200 on success. No request or response body.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## Demoting User from Admin

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `DELETE` | `/api/v2/user/{user_id}/admin` | yes | Admin Only 

Promotes the specified use rto admin. Returns 200 on success. No request or response body.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

## Verifying a User

| Method | Path | Auth Required | Permissions |
| ------ | ---- | ------------- | ----------  |
| `POST` | `/api/v2/user/{user_id}/verify` | yes | Admin Only 

Marks the given user as verified. Returns 200 on success, no request or response body.

__Example Request__: 
```json
```
__Example Response__:
```json
{}
```

