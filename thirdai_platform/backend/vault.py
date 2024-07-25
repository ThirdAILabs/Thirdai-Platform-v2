from fastapi import APIRouter, HTTPException, Depends, status, Header
from sqlalchemy.orm import Session
from database.session import get_session
from auth.jwt import AuthenticatedUser, verify_access_token
from database import schema
from backend.utils import response, get_model_from_identifier
import hvac
from pydantic import BaseModel

vault_app = APIRouter()


def get_vault_client():
    client = hvac.Client(url="http://127.0.0.1:8200", token="YOUR_VAULT_TOKEN")
    if not client.is_authenticated():
        raise HTTPException(status_code=500, detail="Vault authentication failed")
    return client


def verify_admin_access(current_user: schema.User = Depends(get_current_user)):
    if not current_user.admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Admin privileges required"
        )
    return current_user


def verify_model_access(
    model_identifier: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user = authenticated_user.user
    try:
        model: schema.Model = get_model_from_identifier(model_identifier, session)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )
    if model.user_id != user.id and not user.admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Access denied to the model"
        )


def get_current_user(
    session: Session = Depends(get_session),
    user: schema.User = Depends(verify_access_token),
):
    user = session.query(schema.User).filter(schema.User.id == user.user.id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user"
        )
    return user


class SecretRequest(BaseModel):
    user_id: str
    key: str
    value: str


class SecretResponse(BaseModel):
    user_id: str
    key: str
    value: str


@vault_app.post("/add_secret", response_model=SecretResponse)
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


@vault_app.get("/get_secret/{user_id}/{key}", response_model=SecretResponse)
async def get_secret(
    user_id: str, key: str, client: hvac.Client = Depends(get_vault_client)
):
    if key not in ["AWS_ACCESS_TOKEN", "OPENAI_API_KEY"]:
        raise HTTPException(
            status_code=400,
            detail="Invalid key. Only 'AWS_ACCESS_TOKEN' and 'OPENAI_API_KEY' are allowed.",
        )
    secret_path = f"secret/data/{user_id}/{key}"
    read_response = client.secrets.kv.v2.read_secret_version(path=secret_path)
    if (
        "data" not in read_response["data"]
        or "value" not in read_response["data"]["data"]
    ):
        raise HTTPException(status_code=404, detail="Secret not found")
    secret_value = read_response["data"]["data"]["value"]
    return SecretResponse(user_id=user_id, key=key, value=secret_value)
