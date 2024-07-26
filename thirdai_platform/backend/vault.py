from fastapi import APIRouter, HTTPException, Depends, status, Header
from sqlalchemy.orm import Session
from database.session import get_session
from auth.jwt import AuthenticatedUser, verify_access_token
from database import schema
from thirdai_platform.backend.routers.utils import response, get_model_from_identifier
import hvac
from backend.auth_dependencies import verify_admin_access, get_vault_client
from pydantic import BaseModel

vault_router = APIRouter()


class SecretRequest(BaseModel):
    user_id: str
    key: str
    value: str


class SecretResponse(BaseModel):
    user_id: str
    key: str
    value: str


@vault_router.post("/add_secret", response_model=SecretResponse)
async def add_secret(
    secret: SecretRequest,
    client: hvac.Client = Depends(get_vault_client),
    _=Depends(verify_admin_access),
):
    if secret.key not in ["AWS_ACCESS_TOKEN", "OPENAI_API_KEY"]:
        raise HTTPException(
            status_code=400,
            detail="Invalid key. Only 'AWS_ACCESS_TOKEN' and 'OPENAI_API_KEY' are allowed.",
        )
    secret_path = f"secret/data/{secret.user_id}/{secret.key}"
    client.secrets.kv.v2.create_or_update_secret(
        path=secret_path, secret={"value": secret.value}
    )
    return SecretResponse(user_id=secret.user_id, key=secret.key, value=secret.value)


@vault_router.get("/get_secret/{user_id}/{key}", response_model=SecretResponse)
async def get_secret(
    user_id: str, key: str, client: hvac.Client = Depends(get_vault_client)
):
    if key not in ["AWS_ACCESS_TOKEN", "OPENAI_API_KEY"]:
        raise HTTPException(
            status_code=400,
            detail="Invalid key. Only 'AWS_ACCESS_TOKEN' and 'OPENAI_API_KEY' are allowed.",
        )
    secret_path = f"secret/data/{user_id}/{key}"
    try:
        read_response = client.secrets.kv.v2.read_secret_version(path=secret_path)
    except hvac.exceptions.InvalidPath as e:
        return HTTPException(status_code=404, detail="Secret not found")

    secret_value = read_response["data"]["data"]["value"]
    return SecretResponse(user_id=user_id, key=key, value=secret_value)
