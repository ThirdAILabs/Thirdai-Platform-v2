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


keycloak_admin = KeycloakAdmin(
    server_url="http://localhost:8180/",
    username="kc_admin",
    password="password",
    realm_name="master",
    verify=True,
)


def create_client(client_name: str, redirect_uris: list):
    """Create a new confidential client in Keycloak with the necessary permissions."""
    clients = keycloak_admin.get_clients()

    if any(client["clientId"] == client_name for client in clients):
        print(f"Client '{client_name}' already exists.")
        return

    new_client = {
        "clientId": client_name,
        "enabled": True,
        "publicClient": True,
        "redirectUris": redirect_uris,
        "directAccessGrantsEnabled": True,
        "serviceAccountsEnabled": False,
        "standardFlowEnabled": True,
        "implicitFlowEnabled": False,
        "fullScopeAllowed": True,
        "defaultClientScopes": [
            "profile",
            "email",
            "openid",
        ],
        "optionalClientScopes": ["offline_access"],
    }

    keycloak_admin.create_client(new_client)
    print(f"Client '{client_name}' created successfully.")


client_name = "new-client"

create_client(client_name=client_name, redirect_uris=["http://localhost:8180/*"])

keycloak_openid = KeycloakOpenID(
    server_url="http://localhost:8180/",
    client_id=client_name,
    realm_name="master",
)


oauth2_scheme = OAuth2AuthorizationCodeBearer(
    authorizationUrl="http://localhost:8180/realms/master/protocol/openid-connect/auth?client_id=new-client",
    tokenUrl="http://localhost:8180/realms/master/protocol/openid-connect/token",
)


def create_realm_role(role_name: str):
    """Creates realm role in Keycloak if it doesn't exist"""
    roles = keycloak_admin.get_realm_roles()
    if not any(role["name"] == role_name for role in roles):
        keycloak_admin.create_realm_role(payload={"name": role_name})
        print(f"Role '{role_name}' created successfully.")
    else:
        print(f"Role '{role_name}' already exists.")


def initialize_keycloak_roles():
    """Creates the necessary roles in Keycloak: global_admin, team_admin, user."""
    create_realm_role("global_admin")
    create_realm_role("team_admin")
    create_realm_role("user")


initialize_keycloak_roles()


def sync_role_in_keycloak(user_email: str, role_name: str, action: str):
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
