import os
import uuid

import pytest
import requests
from utils import doc_dir

from client.bazaar import ModelBazaar


@pytest.mark.unit
def test_in_memory_deployment():
    base_url = "http://127.0.0.1:80/api/"

    admin_client = ModelBazaar(base_url)
    admin_client.log_in("admin@mail.com", "password")

    base_model_name = f"basic_ndb_{uuid.uuid4()}"
    base_model = admin_client.train(
        base_model_name,
        unsupervised_docs=[os.path.join(doc_dir(), "articles.csv")],
        model_options={"on_disk": False},
        supervised_docs=[],
    )

    with pytest.raises(requests.exceptions.HTTPError, match=".*400.*"):
        ndb_client = admin_client.deploy(
            base_model.model_identifier, autoscaling_enabled=False
        )

    ndb_client = admin_client.deploy(
        base_model.model_identifier, autoscaling_enabled=True
    )

    admin_client.undeploy(ndb_client)
