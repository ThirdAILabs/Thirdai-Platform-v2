import re
from collections import defaultdict
from typing import Dict, List
from urllib.parse import urljoin

from deployment_job.pydantic_models.inputs import PiiEntity
from fastapi import HTTPException, status
from platform_common.logging import JobLogger, LogCode
from requests import Session


def max_overlap(a: str, b: str) -> int:
    def longest_prefix(i: int, j: int) -> int:
        cnt = 0
        for k in range(0, min(len(a) - i, len(b) - j)):
            if a[i + k] == b[j + k]:
                cnt += 1
            else:
                break
        return cnt

    return max(longest_prefix(i, j) for i in range(len(a)) for j in range(len(b)))


def merge_tags(tokens: List[str], tags: List[List[str]]):
    if len(tags) < 1:
        return tokens, tags
    merged_tokens = []
    merged_tags = []

    curr_span = []
    curr_tag = tags[0][0]

    for token, tag in zip(tokens, tags):
        if tag[0] == curr_tag:
            curr_span.append(token)
        else:
            merged_tokens.append(" ".join(curr_span))
            merged_tags.append(curr_tag)
            curr_span = [token]
            curr_tag = tag[0]
    merged_tokens.append(" ".join(curr_span))
    merged_tags.append(curr_tag)

    return merged_tokens, merged_tags


class LabelMap:
    def __init__(self):
        self.tag_to_entities: Dict[str, Dict[str, str]] = defaultdict(dict)
        self.next_label = 0

    def get_label(self, tag: str, entity: str) -> str:
        for label, existing_entity in self.tag_to_entities[tag].items():
            if entity == existing_entity or max_overlap(entity, existing_entity) > 5:
                return label

        label = f"[{tag}#{self.next_label}]"
        self.next_label += 1

        self.tag_to_entities[tag][label] = entity
        return label

    def get_entities(self) -> List[PiiEntity]:
        return [
            PiiEntity(token=token, label=label)
            for labels in self.tag_to_entities.values()
            for label, token in labels.items()
        ]


class Guardrail:
    def __init__(
        self, guardrail_model_id: str, model_bazaar_endpoint: str, logger: JobLogger
    ):
        self.session = Session()
        self.endpoint = urljoin(model_bazaar_endpoint, f"{guardrail_model_id}/predict")
        self.logger = logger

    def query_pii_model(self, text: str, access_token: str):
        res = self.session.post(
            self.endpoint,
            headers={
                "User-Agent": "NDB Deployment job",
                "Authorization": f"Bearer {access_token}",
            },
            json={"text": text, "data_type": "unstructured"},
        )

        if res.status_code != status.HTTP_200_OK:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Unable to access guardrail model: error {res.status_code}",
            )

        results = res.json()["data"]["prediction_results"]

        if results["data_type"] != "unstructured":
            message = f"Guardrail model returned non-unstructured data type: {results['data_type']}"
            self.logger.error(message, code=LogCode.GUARDRAILS)
            raise ValueError(message)

        return results

    def redact_pii(self, text: str, access_token: str, label_map: LabelMap):
        try:
            data = self.query_pii_model(text=text, access_token=access_token)

            entities, tags = merge_tags(
                tokens=data["tokens"], tags=data["predicted_tags"]
            )

            entities = [
                label_map.get_label(tag=tag, entity=entity) if tag != "O" else entity
                for entity, tag in zip(entities, tags)
            ]

            self.logger.debug(f"Redacted PII: {entities}", code=LogCode.GUARDRAILS)

            return " ".join(entities)
        except Exception as e:
            message = f"Error redacting PII: {e}"
            self.logger.error(message, code=LogCode.GUARDRAILS)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=message
            )

    def unredact_pii(self, redacted_text: str, entities: List[PiiEntity]):
        try:
            entity_map = {entity.label: entity.token for entity in entities}

            def replace(match):
                return entity_map.get(match[0], "[UNKNOWN ENTITY]")

            return re.sub(r"\[([A-Z]+)#(\d+)\]", replace, redacted_text)
        except Exception as e:
            message = f"Error unredacting PII: {e}"
            self.logger.error(message, code=LogCode.GUARDRAILS)
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=message
            )
