from typing import Dict, List, Optional

from auth.jwt import AuthenticatedUser, verify_access_token
from database.session import get_session
from fastapi import APIRouter, Depends, status
from pydantic import BaseModel
from sqlalchemy.orm import Session
from utils import response

data_router = APIRouter()


class TextClassificationGenerateArgs(BaseModel):
    samples_per_label: int
    target_labels: List[str]
    user_vocab: Optional[List[str]] = []
    examples: Optional[Dict[str, List[str]]] = None
    user_prompts: Optional[List[str]] = None
    labels_description: Optional[Dict[str, str]] = None
    batch_size: int = 40
    vocab_per_sentence: int = 4


@data_router.post("/generate-text-data")
def generate_text_data(
    task_prompt: str,
    args: TextClassificationGenerateArgs,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    # TODO(Gautam): Only people from ThirdAI should be able to access this endpoint
    try:
        pass
        # TODO(Gautam): Submit data_generation nomad job
    except Exception as e:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, message=str(e)
        )


class TokenClassificationGenerateArgs(BaseModel):
    domain_prompt: str
    tags: List[str]
    tag_examples: Dict[str, List[str]]
    num_call_batches: int
    batch_size: int = 40
    num_samples_per_tag: int = 4


@data_router.post("/generate-token-data")
def generate_text_data(
    task_prompt: str,
    args: TextClassificationGenerateArgs,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    # TODO(Gautam): Only people from ThirdAI should be able to access this endpoint
    try:
        pass
        # TODO(Gautam): Submit data_generation nomad job
    except Exception as e:
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, message=str(e)
        )
