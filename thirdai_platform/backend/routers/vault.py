from fastapi import APIRouter, HTTPException, Depends, status, Header
from sqlalchemy.orm import Session
from database.session import get_session
from auth.jwt import AuthenticatedUser, verify_access_token
from database import schema
from thirdai_platform.backend.routers.utils import response, get_model_from_identifier
import hvac
from backend.auth_dependencies import global_admin_only, get_vault_client
from pydantic import BaseModel

vault_router = APIRouter()


class SecretRequest(BaseModel):
    key: str


class SecretResponse(BaseModel):
    key: str
    value: str


# Note(pratik): Only global admin can add a secret to vault
@vault_router.post(
    "/add-secret",
    dependencies=[Depends(global_admin_only)],
)
async def add_secret(
    secret: SecretResponse,
    client: hvac.Client = Depends(get_vault_client),
):
    if secret.key not in ["AWS_ACCESS_TOKEN", "OPENAI_API_KEY"]:
        raise HTTPException(
            status_code=400,
            detail="Invalid key. Only 'AWS_ACCESS_TOKEN' and 'OPENAI_API_KEY' are allowed.",
        )
    secret_path = f"secret/data/{secret.key}"
    client.secrets.kv.v2.create_or_update_secret(
        path=secret_path, secret={"value": secret.value}
    )
    return {
        "key": secret.key,
        "value": secret.value,
        "status": "success",
    }


# Note(pratik): Any user can access the secrets, set by global admin
@vault_router.get("/get-secret")
async def get_secret(
    secret: SecretRequest, client: hvac.Client = Depends(get_vault_client)
):
    if secret.key not in [
        "AWS_ACCESS_TOKEN",
        "AWS_SECRET_ACCESS_TOKEN",
        "OPENAI_API_KEY",
    ]:
        raise HTTPException(
            status_code=400,
            detail="Invalid key. Only 'AWS_ACCESS_TOKEN' and 'OPENAI_API_KEY' are allowed.",
        )
    secret_path = f"secret/data/{secret.key}"
    try:
        read_response = client.secrets.kv.v2.read_secret_version(path=secret_path)
    except hvac.exceptions.InvalidPath as e:
        return HTTPException(status_code=404, detail="Secret not found")

    secret_value = read_response["data"]["data"]["value"]
    return {
        "key": secret.key,
        "value": secret_value,
        "status": "success",
    }
