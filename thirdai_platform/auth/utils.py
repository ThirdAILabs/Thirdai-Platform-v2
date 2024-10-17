import os
from urllib.parse import urlparse
import socket
import fastapi

socket.setdefaulttimeout(5)


def get_ip_from_url(url):
    try:
        parsed_url = urlparse(url)
        hostname = parsed_url.hostname

        ip_address = socket.gethostbyname(hostname)

        return ip_address
    except Exception as e:
        return f"Error parsing URL or resolving IP: {e}"


identity_provider = os.getenv("IDENTITY_PROVIDER", "postgres")

keycloak_openid = None
keycloak_admin = None
thirdai_realm = "ThirdAI-Platform"

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
    KEYCLOAK_SERVER_URL = os.getenv("KEYCLOAK_SERVER_URL")

    if not KEYCLOAK_SERVER_URL:
        raise ValueError(
            "The environment variable 'KEYCLOAK_SERVER_URL' is required but was not found."
        )

    from keycloak import KeycloakOpenID, KeycloakAdmin

    USE_SSL_IN_LOGIN = os.getenv("USE_SSL_IN_LOGIN", "False").lower() == "true"

    if USE_SSL_IN_LOGIN:
        # Refer: https://github.com/marcospereirampj/python-keycloak/blob/7cfad72a68346ca4cbcba69f9e8808091ae47daa/src/keycloak/keycloak_admin.py#L47
        keycloak_admin = KeycloakAdmin(
            server_url=KEYCLOAK_SERVER_URL,
            username=os.getenv("KEYCLOAK_ADMIN_USER", "temp_admin"),
            password=os.getenv("KEYCLOAK_ADMIN_PASSWORD", "password"),
            realm_name="master",
            verify="/model_bazaar/certs/traefik.crt",
            cert=(
                "/model_bazaar/certs/traefik.crt",
                "/model_bazaar/certs/traefik.key",
            ),
        )
    else:
        # Refer: https://github.com/marcospereirampj/python-keycloak/blob/7cfad72a68346ca4cbcba69f9e8808091ae47daa/src/keycloak/keycloak_admin.py#L47
        keycloak_admin = KeycloakAdmin(
            server_url=KEYCLOAK_SERVER_URL,
            username=os.getenv("KEYCLOAK_ADMIN_USER", "temp_admin"),
            password=os.getenv("KEYCLOAK_ADMIN_PASSWORD", "password"),
            realm_name="master",
            verify=False,
        )

    admin_username = os.getenv("ADMIN_USERNAME")
    admin_mail = os.getenv("ADMIN_MAIL")
    admin_password = os.getenv("ADMIN_PASSWORD")

    keycloak_user_id = keycloak_admin.get_user_id(admin_username)
    # create permanent admin in master 'realm'
    # we create another user with same
    if keycloak_user_id:
        keycloak_admin.update_user(
            keycloak_user_id, {"email": admin_mail, "emailVerified": True}
        )
    else:
        keycloak_user_id = keycloak_admin.create_user(
            {
                "username": admin_username,
                "email": admin_mail,
                "enabled": True,
                "emailVerified": True,
                "credentials": [
                    {
                        "type": "password",
                        "value": admin_password,
                        "temporary": False,
                    }
                ],
            }
        )

    keycloak_roles = keycloak_admin.get_realm_roles()
    role = next((r for r in keycloak_roles if r["name"] == "admin"), None)

    keycloak_admin.assign_realm_roles(keycloak_user_id, [role])

    def create_realm(realm_name: str):
        """Create a new realm in Keycloak."""
        # Refer: https://www.keycloak.org/docs-api/21.1.1/rest-api/#_realmrepresentation
        payload = {
            "realm": realm_name,
            "enabled": True,
            "identityProviders": [],
            "defaultRoles": ["user"],
            "registrationAllowed": True,
            "resetPasswordAllowed": True,
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

    new_realm_name = create_realm(realm_name=thirdai_realm)

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
            # Refer: https://www.keycloak.org/docs-api/21.1.1/rest-api/#_clientrepresentation
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
                ],
            }

            keycloak_admin.create_client(new_client)
            print(f"Client '{client_name}' created successfully.")
        else:
            print(f"Client '{client_name}' already exists.")

    client_name = "thirdai-login-client"
    public_ip = get_ip_from_url(os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT"))
    private_ip = get_ip_from_url(os.getenv("PRIVATE_MODEL_BAZAAR_ENDPOINT"))
    create_client(
        client_name=client_name,
        root_url=KEYCLOAK_SERVER_URL,
        base_url="/login",
        redirect_uris=[
            f"http://{public_ip}/*",
            f"https://{public_ip}/*",
            f"http://{public_ip}:80/*",
            f"https://{public_ip}:80/*",
            f"http://{private_ip}/*",
            f"https://{private_ip}/*",
            f"http://{private_ip}:80/*",
            f"https://{private_ip}:80/*",
            f"http://localhost/*",
            f"https://localhost/*",
            f"http://localhost:80/*",
            f"https://localhost:80/*",
        ],
    )

    keycloak_openid = KeycloakOpenID(
        server_url=KEYCLOAK_SERVER_URL,
        client_id=client_name,
        realm_name=new_realm_name,
    )

    if USE_SSL_IN_LOGIN:
        # Refer: https://github.com/marcospereirampj/python-keycloak/blob/7cfad72a68346ca4cbcba69f9e8808091ae47daa/src/keycloak/keycloak_openid.py#L65
        keycloak_openid = KeycloakOpenID(
            server_url=KEYCLOAK_SERVER_URL,
            client_id=client_name,
            realm_name=new_realm_name,
            verify="/model_bazaar/certs/traefik.crt",
            cert=(
                "/model_bazaar/certs/traefik.crt",
                "/model_bazaar/certs/traefik.key",
            ),
        )
    else:
        # Refer: https://github.com/marcospereirampj/python-keycloak/blob/7cfad72a68346ca4cbcba69f9e8808091ae47daa/src/keycloak/keycloak_openid.py#L65
        keycloak_openid = KeycloakOpenID(
            server_url=KEYCLOAK_SERVER_URL,
            client_id=client_name,
            realm_name=new_realm_name,
        )
