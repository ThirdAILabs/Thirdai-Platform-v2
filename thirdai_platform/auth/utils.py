import os
import fastapi
import requests
from keycloak import KeycloakOpenID, KeycloakAdmin
from fastapi.security import OAuth2AuthorizationCodeBearer, OAuth2PasswordBearer
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

IDENTITY_PROVIDER = os.getenv("IDENTITY_PROVIDER", "postgres").lower()


if IDENTITY_PROVIDER == "keycloak":
    # Keycloak admin client initialization
    keycloak_admin = KeycloakAdmin(
        server_url="http://localhost:8180/",
        username=os.getenv("KEYCLOAK_ADMIN_USER", "kc_admin"),
        password=os.getenv("KEYCLOAK_ADMIN_PASSWORD", "password"),
        realm_name="master",
        verify=True,  # Optional: False if we are skipping SSL verification
    )

    def create_realm(realm_name: str):
        """Create a new realm in Keycloak."""
        payload = {
            "realm": realm_name,
            "enabled": True,
            "sslRequired": "None",
            "identityProviders": [],
            "defaultRoles": ["user"],
        }

        current_realms = [
            realms_metadata["realm"] for realms_metadata in keycloak_admin.get_realms()
        ]
        return realm_name
        if realm_name in current_realms:
            keycloak_admin.delete_realm(realm_name)

        try:
            response = keycloak_admin.create_realm(payload)
            print(f"Realm '{realm_name}' created successfully: {response}")
            return realm_name  # Return the created realm name
        except Exception as e:
            print(f"Error creating realm '{realm_name}': {str(e)}")
            return None

    new_realm_name = create_realm(realm_name="new-realm")

    keycloak_admin.change_current_realm(new_realm_name)

    def create_client(
        client_name: str, redirect_uris: list, root_url: str, base_url: str
    ):
        """Create a new confidential client in Keycloak with the necessary permissions."""
        clients = keycloak_admin.get_clients()
        existing_client = next(
            (client for client in clients if client["clientId"] == client_name), None
        )

        if existing_client:
            return
            # keycloak_admin.delete_client(existing_client["id"])
            # print(f"Client '{client_name}' already existed and was deleted.")

        new_client = {
            "clientId": client_name,
            "enabled": True,
            "publicClient": True,  # Set to False to use client-secret authentication.
            "redirectUris": redirect_uris,
            "rootUrl": root_url,
            "baseUrl": base_url,
            "directAccessGrantsEnabled": False,  # Align with account-console if direct grants are not needed.
            "serviceAccountsEnabled": False,
            "standardFlowEnabled": True,
            "implicitFlowEnabled": False,
            "fullScopeAllowed": False,  # Adjust to match account-console settings.
            "defaultClientScopes": ["profile", "email", "openid", "roles"],
            "optionalClientScopes": ["offline_access", "microprofile-jwt"],
            "protocolMappers": [
                {
                    "name": "audience resolve",
                    "protocol": "openid-connect",
                    "protocolMapper": "oidc-audience-resolve-mapper",
                    "consentRequired": False,
                    "config": {},
                }
            ],
            "webOrigins": ["*", "http://localhost:80/*", "http://localhost:8180/*"],
        }

        keycloak_admin.create_client(new_client)
        print(f"Client '{client_name}' created successfully.")

    client_name = "new-client"
    # create_client(
    #     client_name=client_name,
    #     root_url="http://localhost:8180",
    #     base_url="/login",
    #     redirect_uris=["http://localhost:8180/*", "http://localhost:80/*"],
    # )

    keycloak_openid = KeycloakOpenID(
        server_url="http://localhost:8180/",
        client_id=client_name,
        realm_name=new_realm_name,
    )

    oauth2_scheme = OAuth2AuthorizationCodeBearer(
        authorizationUrl=f"http://localhost:8180/realms/{new_realm_name}/protocol/openid-connect/auth?client_id={client_name}",
        tokenUrl=f"http://localhost:8180/realms/{new_realm_name}/protocol/openid-connect/token",
    )

    def create_realm_role(role_name: str):
        """Create realm role in Keycloak if it doesn't exist."""
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

    def get_token(client_id, username, password):
        """Obtain an access token from Keycloak."""
        data = {
            "grant_type": "password",
            "client_id": client_id,
            "username": username,
            "password": password,
            "scope": "openid profile email",  # Scopes required for OIDC
        }

        response = requests.post(
            f"http://localhost:8180/realms/{new_realm_name}/protocol/openid-connect/token",
            data=data,
            headers={"Content-Type": "application/x-www-form-urlencoded"},
        )

        if response.status_code == 200:
            return response.json()["access_token"]  # Return access token
        else:
            print(f"Failed to obtain token: {response.status_code} {response.text}")
            return None

else:
    # If Keycloak is not used, define these as placeholders
    keycloak_openid = None
    keycloak_admin = None
    oauth2_scheme = None

    def create_client(*args, **kwargs):
        pass

    def initialize_keycloak_roles():
        pass

    def get_token(*args, **kwargs):
        return None
