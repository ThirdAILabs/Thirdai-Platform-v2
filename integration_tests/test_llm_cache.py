import glob

pass
import os
import uuid

import pytest
import requests
from utils import doc_dir

from client.bazaar import ModelBazaar


@pytest.mark.unit
def test_llm_cache():
    base_url = "http://127.0.0.1:80/api/"

    admin_client = ModelBazaar(base_url)
    admin_client.log_in("admin@mail.com", "password")

    base_model_name = f"basic_ndb_{uuid.uuid4()}"
    base_model = admin_client.train(
        base_model_name,
        unsupervised_docs=[os.path.join(doc_dir(), "articles.csv")],
        supervised_docs=[],
    )

    ndb_client = admin_client.deploy(
        base_model.model_identifier, autoscaling_enabled=True
    )

    def auth_header():
        return {
            "Authorization": f"Bearer {admin_client._access_token}",
        }

    model_id = base_model.model_id

    cache_health_url = f"http://127.0.0.1:80/{model_id}/cache/health"
    response = requests.get(cache_health_url)
    assert response.status_code == 200

    model_bazaar_dir = os.getenv("SHARE_DIR")
    cache_file_path = os.path.join(
        model_bazaar_dir, "models", model_id, "llm_cache", "llm_cache.ndb"
    )
    assert os.path.exists(cache_file_path)

    suggestions_url = f"http://127.0.0.1:80/{model_id}/cache/suggestions"
    suggestions_response = requests.get(
        suggestions_url,
        params={"query": "lol", "model_id": model_id},
        headers=auth_header(),
    )
    assert suggestions_response.status_code == 200
    assert not suggestions_response.json()["suggestions"]

    cache_token_url = f"http://127.0.0.1:80/{model_id}/cache/token"
    token_response = requests.get(
        cache_token_url, params={"model_id": model_id}, headers=auth_header()
    )
    assert token_response.status_code == 200
    cache_token = token_response.json()["access_token"]
    assert cache_token

    insertion_url = f"http://127.0.0.1:80/{model_id}/cache/insert"
    insertion_response = requests.post(
        insertion_url,
        json={"query": "lol", "llm_res": "response", "references": ["something"]},
        headers={"Authorization": f"Bearer {cache_token}"},
    )
    insertion_response

    pattern = os.path.join(
        model_bazaar_dir, "models", model_id, "llm_cache", "insertions", "*.jsonl"
    )
    matching_files = glob.glob(pattern)
    assert len(matching_files) > 0

    admin_client.undeploy(ndb_client)

    stopped_response = requests.get(cache_health_url)
    assert stopped_response.status_code != 200

    refresh_url = f"{base_url}train/refresh-llm-cache"
    refresh_response = requests.post(
        refresh_url,
        params={"model_identifier": base_model.model_identifier},
        headers=auth_header(),
    )
    assert refresh_response.status_code == 200

    pattern = os.path.join(
        model_bazaar_dir, "models", model_id, "llm_cache", "insertions", "*.jsonl"
    )
    matching_files = glob.glob(pattern)
    assert len(matching_files) == 0

    ndb_client = admin_client.deploy(
        base_model.model_identifier, autoscaling_enabled=True
    )

    suggestions_response_after_restart = requests.get(
        suggestions_url,
        params={"query": "lol", "model_id": model_id},
        headers=auth_header(),
    )
    assert suggestions_response_after_restart.status_code != 200

    admin_client.undeploy(ndb_client)
