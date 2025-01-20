import os
import uuid
from typing import Any, Dict, Literal, Optional

from platform_common.pydantic_models.training import ModelType
from pydantic import BaseModel, Field


class KnowledgeExtractionOptions(BaseModel):
    model_type: Literal[ModelType.KNOWLEDGE_EXTRACTION] = ModelType.KNOWLEDGE_EXTRACTION

    llm_provider: str = "openai"
    genai_key: Optional[str] = None

    advanced_indexing: bool
    rerank: bool
    generate_answers: bool


class DeploymentConfig(BaseModel):
    deployment_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    user_id: str
    model_id: str
    model_type: ModelType
    model_bazaar_endpoint: str
    model_bazaar_dir: str
    host_dir: str
    license_key: str
    job_auth_token: str

    autoscaling_enabled: bool = False

    options: Dict[str, Any]

    class Config:
        protected_namespaces = ()

    def get_nomad_endpoint(self) -> str:
        # Parse the model_bazaar_endpoint to extract scheme and host
        from urllib.parse import urlparse, urlunparse

        parsed_url = urlparse(self.model_bazaar_endpoint)

        # Reconstruct the URL with port 4646
        nomad_netloc = f"{parsed_url.hostname}:4646"

        # Rebuild the URL while keeping the original scheme and hostname
        return urlunparse((parsed_url.scheme, nomad_netloc, "", "", "", ""))

    def save_deployment_config(self):
        config_path = os.path.join(
            self.model_bazaar_dir,
            "models",
            str(self.model_id),
            "deployments",
            "deployment_config.json",
        )
        os.makedirs(os.path.dirname(config_path), exist_ok=True)
        with open(config_path, "w") as file:
            file.write(self.model_dump_json(indent=4))

        return config_path
