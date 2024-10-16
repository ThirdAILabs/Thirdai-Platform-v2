import os
import re
from typing import Literal, Optional, Union

from platform_common.pydantic_models.training import ModelType, NDBSubType, UDTSubType
from pydantic import BaseModel, Field


class NDBDeploymentOptions(BaseModel):
    model_type: Literal[ModelType.NDB] = ModelType.NDB

    ndb_sub_type: NDBSubType = NDBSubType.v2

    llm_provider: str = "openai"
    genai_key: Optional[str] = None


class UDTDeploymentOptions(BaseModel):
    model_type: Literal[ModelType.UDT] = ModelType.UDT

    udt_sub_type: UDTSubType


class DeploymentConfig(BaseModel):
    model_id: str
    model_bazaar_endpoint: str
    model_bazaar_dir: str
    license_key: str

    autoscaling_enabled: bool = False

    model_options: Union[NDBDeploymentOptions, UDTDeploymentOptions] = Field(
        ..., discriminator="model_type"
    )

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


class UsageStatOptions(BaseModel):
    duration: int  # In seconds
    step: int  # 1h, 20s, 3m, 2w, 1h30m or any other prometheus supported duration regex (https://prometheus.io/docs/prometheus/latest/configuration/configuration/#:~:text=(((%5B0%2D9%5D%2B)y)%3F((%5B0%2D9%5D%2B)w)%3F((%5B0%2D9%5D%2B)d)%3F((%5B0%2D9%5D%2B)h)%3F((%5B0%2D9%5D%2B)m)%3F((%5B0%2D9%5D%2B)s)%3F((%5B0%2D9%5D%2B)ms)%3F%7C0))

    def step_in_words(self):
        unit_map = {
            "y": "year",
            "w": "week",
            "d": "day",
            "h": "hour",
            "m": "minute",
            "s": "second",
            "ms": "millisecond",
        }

        pattern = r"(\d+)([ywdhms]+|ms)"
        matches = re.findall(pattern, self.step)

        parts = []
        for value, unit in matches:
            value = int(value)
            if unit in unit_map:
                unit_name = unit_map[unit]
                if value != 1:
                    unit_name = f"{unit_name}s"
                parts.append(f"{value} {unit_name}")

        return f"per {' '.join(parts)}"
