import os
import fastapi
import requests
from fastapi.security import OAuth2AuthorizationCodeBearer, OAuth2PasswordBearer
from fastapi import APIRouter, Depends, HTTPException, status

keycloak_openid = None
keycloak_admin = None

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
    from keycloak import KeycloakOpenID, KeycloakAdmin

    keycloak_admin = KeycloakAdmin(
        server_url="http://localhost:8180/",
        username=os.getenv("KEYCLOAK_ADMIN_USER", "kc_admin"),
        password=os.getenv("KEYCLOAK_ADMIN_PASSWORD", "password"),
        realm_name="master",
        verify=True,
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

        if realm_name not in current_realms:
            try:
                response = keycloak_admin.create_realm(payload)
                print(f"Realm '{realm_name}' created successfully: {response}")
            except Exception as e:
                print(f"Error creating realm '{realm_name}': {str(e)}")
                return None

        return realm_name

    new_realm_name = create_realm(realm_name="ThirdAI-Platform")

    keycloak_admin.change_current_realm(new_realm_name)

    def create_client(
        client_name: str, redirect_uris: list, root_url: str, base_url: str
    ):
        """Create a new confidential client in Keycloak with the necessary permissions."""
        clients = keycloak_admin.get_clients()
        existing_client = next(
            (client for client in clients if client["clientId"] == client_name), None
        )

        if not existing_client:
            new_client = {
                "clientId": client_name,
                "enabled": True,
                "publicClient": True,
                "redirectUris": redirect_uris,
                "rootUrl": root_url,
                "baseUrl": base_url,
                "directAccessGrantsEnabled": False,
                "serviceAccountsEnabled": False,
                "standardFlowEnabled": True,
                "implicitFlowEnabled": False,
                "fullScopeAllowed": False,
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
                "webOrigins": [
                    "*",
                    "http://localhost:80/*",
                    "http://localhost:8180/*",
                    "http://localhost/*",
                ],
            }

            keycloak_admin.create_client(new_client)
            print(f"Client '{client_name}' created successfully.")
        else:
            print(f"Client '{client_name}' already exists.")

    client_name = "thirdai-login-client"
    create_client(
        client_name=client_name,
        root_url="http://localhost:8180",
        base_url="/login",
        redirect_uris=[
            "http://localhost:8180/*",
            "http://localhost:80/*",
            "http://localhost:3006/*",
            "http://localhost/*",
        ],
    )

    keycloak_openid = KeycloakOpenID(
        server_url="http://localhost:8180/",
        client_id=client_name,
        realm_name=new_realm_name,
    )
