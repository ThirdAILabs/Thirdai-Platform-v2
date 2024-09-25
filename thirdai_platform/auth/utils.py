import fastapi

from keycloak import KeycloakOpenID, KeycloakAdmin
from fastapi.security import OAuth2AuthorizationCodeBearer
from fastapi import APIRouter, Depends, HTTPException, status

# This is a helper object provided by FastAPI to extract the access token from a
# request. The paramter 'tokenURL' specifies which endpoint the user will use in
# the app to obtain the access token. When this is used in the fastapi.Depends
# clause in an endpoint (see `verify_access_token` below) it will find the
# 'Authorization' header in the request and check that the value of the header is
# 'Bearer <access token>' and then return the value of the access token within the
# header. It will return an 401 unauthorized code if the header is not found or
# is not properly formatted.
token_bearer = fastapi.security.OAuth2PasswordBearer(tokenUrl="/auth/token")

CREDENTIALS_EXCEPTION = fastapi.HTTPException(
    status_code=fastapi.status.HTTP_401_UNAUTHORIZED,
    detail="Invalid access token.",
    # This header indicates what type of authentication would be required to access
    # the resource.
    # https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
    headers={"WWW-Authenticate": "Bearer"},
)

keycloak_openid = KeycloakOpenID(
    server_url="http://localhost:8180/auth/",
    client_id="myclient",
    realm_name="myrealm",
    client_secret_key="myclientsecret",
)


keycloak_admin = KeycloakAdmin(
    server_url="http://localhost:8080/auth/",
    username="admin",
    password="admin_password",
    realm_name="myrealm",
    client_id="admin-cli",
    verify=True,
)

oauth2_scheme = OAuth2AuthorizationCodeBearer(
    authorizationUrl="http://localhost:8180/auth/realms/myrealm/protocol/openid-connect/auth",
    tokenUrl="http://localhost:8180/auth/realms/myrealm/protocol/openid-connect/token",
)


def sync_role_in_keycloak(user_email: str, role_name: str, action: str):
    """
    Sync the role of a user in Keycloak by adding or removing the role.

    Parameters:
    - user_email: The email of the user.
    - role_name: The role name to assign or remove.
    - action: "add" to add the role or "remove" to remove the role.
    """
    keycloak_user = keycloak_admin.get_user_by_email(user_email)
    if not keycloak_user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found in Keycloak"
        )

    keycloak_roles = keycloak_admin.get_realm_roles()
    role = next((r for r in keycloak_roles if r["name"] == role_name), None)
    if not role:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Role not found in Keycloak"
        )

    if action == "add":
        keycloak_admin.assign_realm_roles(user_id=keycloak_user["id"], roles=[role])
    elif action == "remove":
        keycloak_admin.remove_realm_roles(user_id=keycloak_user["id"], roles=[role])
    else:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, detail="Invalid action"
        )
