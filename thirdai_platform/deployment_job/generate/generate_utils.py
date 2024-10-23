import logging
import os
from typing import List, Optional
from urllib.parse import urljoin

import requests
from deployment_job.pydantic_models.inputs import Reference

DEFAULT_SYSTEM_PROMPT = (
    "Write a short answer for the user's query based on the provided context. "
    "If the context provides insufficient information, mention it but answer to "
    "the best of your abilities."
)

DEFAULT_PROMPT = "Given this context, "


def reference_content(reference: Reference) -> str:
    if (
        reference.source
        and reference.source.endswith(".pdf")
        or reference.source.endswith(".docx")
    ):
        return f'(From file "{reference.source}") {reference.text}'
    if "title" in reference.metadata:
        return f'(From file "{reference.metadata["title"]}") {reference.text}'
    return f"(From a webpage) {reference.text}"


def make_prompt(
    query: str,
    task_prompt: Optional[str],
    references: List[Reference],
    reverse_ref_order: bool = False,
    token_limit: int = 2000,
):
    if reverse_ref_order:
        references = references[::-1]

    processed_references = map(reference_content, references)
    context = "\n\n".join(processed_references)

    if reverse_ref_order:
        context = " ".join(context.split(" ")[-token_limit:])
    else:
        context = " ".join(context.split(" ")[:token_limit])

    system_prompt = DEFAULT_SYSTEM_PROMPT
    user_prompt = f"{context}\n\n {task_prompt or DEFAULT_PROMPT} {query}"

    return system_prompt, user_prompt


async def insert_into_cache(
    original_query: str, generated_response: str, cache_access_token: str
):
    try:
        res = requests.post(
            urljoin(os.environ["MODEL_BAZAAR_ENDPOINT"], "/cache/insert"),
            params={
                "query": original_query,
                "llm_res": generated_response,
            },
            headers={
                "Authorization": f"Bearer {cache_access_token}",
            },
        )
        if res.status_code != 200:
            logging.error(f"LLM Cache Insertion failed: {res}")
    except Exception as e:
        logging.error("LLM Cache Insert Error", e)
