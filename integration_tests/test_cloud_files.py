import os
import uuid

import pytest

from client.bazaar import ModelBazaar


@pytest.mark.parametrize(
    "model_name_prefix, doc_url, provider, expected_query",
    [
        (
            "s3_public_ndb",
            "s3://thirdai-corp-public/ThirdAI-Enterprise-Test-Data/scifact/",
            "s3",
            "sample query",
        ),
        (
            "s3_private_ndb",
            "s3://thirdai-datasets/insert.pdf",
            "s3",
            "Alice in wonderland",
        ),
        # (
        #     "azure_public_ndb",
        #     "https://csg100320028d93f3bc.blob.core.windows.net/test/insert.pdf",
        #     "azure",
        #     "Alice in wonderland",
        # ),
        # (
        #     "azure_private_ndb",
        #     "https://csg100320028d93f3bc.blob.core.windows.net/private-platform/insert.pdf",
        #     "azure",
        #     "Alice in wonderland",
        # ),
        # (
        #     "gcp_public_ndb",
        #     "gs://public-training-platform/sample_nda.pdf",
        #     "gcp",
        #     "confidentiality agreement",
        # ),
        # (
        #     "gcp_private_ndb",
        #     "gs://private-thirdai-platform/sample_nda.pdf",
        #     "gcp",
        #     "confidentiality agreement",
        # ),
    ],
)
@pytest.mark.unit
def test_cloud_training(model_name_prefix, doc_url, provider, expected_query):
    base_url = "http://127.0.0.1:80/api/"
    admin_client = ModelBazaar(base_url)
    admin_client.log_in("admin@mail.com", "password")

    # Dynamically generate the model name based on the prefix and uuid
    model_name = f"{model_name_prefix}_{uuid.uuid4()}"

    # Train the model with the corresponding file URL and provider
    model = admin_client.train(
        model_name,
        unsupervised_docs=[doc_url],
        model_options={"ndb_options": {"ndb_sub_type": "v2"}},
        supervised_docs=[],
        doc_type=provider,
    )
    admin_client.await_train(model)

    # Deploy model and validate results
    ndb_client = admin_client.deploy(model.model_identifier)
    admin_client.await_deploy(ndb_client)

    # Search and validate
    res = ndb_client.search(expected_query, top_k=1)
    assert res["references"][0]["id"] is not None

    # Get signed URL and check access
    signed_url = ndb_client.get_signed_url(
        source=res["references"][0]["source"], provider=provider
    )
    assert signed_url is not None

    # Validate the signed URL using wget or another method
    response = os.system(f"wget -q --spider {signed_url}")
    assert response == 0, f"Failed to access {provider} signed URL: {signed_url}"

    # Undeploy the model after validation
    admin_client.undeploy(ndb_client)
