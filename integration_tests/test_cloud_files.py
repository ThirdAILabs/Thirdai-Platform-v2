import uuid

import pytest

from client.bazaar import ModelBazaar


@pytest.mark.unit
def test_s3_training():
    base_url = "http://127.0.0.1:80/api/"
    admin_client = ModelBazaar(base_url)
    admin_client.log_in("admin@mail.com", "password")

    model_name = f"s3_ndb_{uuid.uuid4()}"
    model = admin_client.train(
        model_name,
        unsupervised_docs=[
            "s3://thirdai-corp-public/ThirdAI-Enterprise-Test-Data/scifact/"
        ],
        model_options={"ndb_options": {"ndb_sub_type": "v2"}},
        supervised_docs=[],
        doc_type="s3",
    )
    admin_client.await_train(model)

    # Deploy model and validate results
    ndb_client = admin_client.deploy(model.model_identifier)
    admin_client.await_deploy(ndb_client)

    # Search and validate
    res = ndb_client.search("sample query", top_k=1)
    assert res["references"][0]["id"] is not None

    admin_client.undeploy(ndb_client)


@pytest.mark.unit
def test_azure_training():
    base_url = "http://127.0.0.1:80/api/"
    admin_client = ModelBazaar(base_url)
    admin_client.log_in("admin@mail.com", "password")

    model_name = f"azure_ndb_{uuid.uuid4()}"
    model = admin_client.train(
        model_name,
        unsupervised_docs=[
            "https://csg100320028d93f3bc.blob.core.windows.net/test/insert.pdf"
        ],
        model_options={"ndb_options": {"ndb_sub_type": "v2"}},
        supervised_docs=[],
        doc_type="azure",
    )
    admin_client.await_train(model)

    # Deploy model and validate results
    ndb_client = admin_client.deploy(model.model_identifier)
    admin_client.await_deploy(ndb_client)

    # Search and validate
    res = ndb_client.search("sample query", top_k=1)
    assert res["references"][0]["id"] is not None

    admin_client.undeploy(ndb_client)


@pytest.mark.unit
def test_gcp_training():
    base_url = "http://127.0.0.1:80/api/"
    admin_client = ModelBazaar(base_url)
    admin_client.log_in("admin@mail.com", "password")

    model_name = f"gcp_ndb_{uuid.uuid4()}"
    model = admin_client.train(
        model_name,
        unsupervised_docs=["gs://public-training-platform/sample_nda.pdf"],
        model_options={"ndb_options": {"ndb_sub_type": "v2"}},
        supervised_docs=[],
        doc_type="gcp",
    )
    admin_client.await_train(model)

    # Deploy model and validate results
    ndb_client = admin_client.deploy(model.model_identifier)
    admin_client.await_deploy(ndb_client)

    # Search and validate
    res = ndb_client.search("sample query", top_k=1)
    assert res["references"][0]["id"] is not None

    admin_client.undeploy(ndb_client)
