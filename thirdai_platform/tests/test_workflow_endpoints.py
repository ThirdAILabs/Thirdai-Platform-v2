import pytest
from fastapi.testclient import TestClient

from .utils import auth_header, create_user, login, upload_model

pytestmark = [pytest.mark.unit]

### Setup: Create Users ###


@pytest.fixture(scope="module", autouse=True)
def setup_users():
    """
    Setup fixture to create necessary users for the tests.
    Runs once before all tests in the module.
    """
    from main import app

    client = TestClient(app)
    # Create workflow owner
    res = create_user(
        client, "workflow-owner", "workflow-owner@mail.com", "owner-password"
    )
    assert res.status_code == 200

    # Create normal user
    res = create_user(client, "normal-user", "normal-user@mail.com", "normal-password")
    assert res.status_code == 200


### 1. Workflow Type Tests ###


def test_create_and_delete_workflow_type():
    from main import app

    client = TestClient(app)

    # Log in as global admin
    res = login(client, username="admin@mail.com", password="password")
    assert res.status_code == 200
    admin_jwt = res.json()["data"]["access_token"]

    # Create a complex workflow type
    res = client.post(
        "/api/workflow/add-type",
        json={
            "name": "complex_workflow_type",
            "description": "A complex workflow type",
            "model_requirements": [
                [{"component": "search", "type": "ndb"}],
                [
                    {"component": "filter", "type": "udf", "subtype": "regex"},
                    {"component": "guardrail", "type": "udt", "subtype": "token"},
                ],
            ],
        },
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 200

    # Create a sample workflow type
    res = client.post(
        "/api/workflow/add-type",
        json={
            "name": "sample_workflow_type",
            "description": "A sample workflow type",
            "model_requirements": [[{"component": "search", "type": "ndb"}]],
        },
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 200

    # Access control check: Normal user trying to create workflow type
    res = login(client, username="normal-user@mail.com", password="normal-password")
    assert res.status_code == 200
    normal_user_jwt = res.json()["data"]["access_token"]

    res = client.post(
        "/api/workflow/add-type",
        json={
            "name": "unauthorized_workflow_type",
            "description": "Should not be created",
            "model_requirements": [[{"component": "search", "type": "ndb"}]],
        },
        headers=auth_header(normal_user_jwt),
    )
    assert res.status_code == 403  # Forbidden

    # Attempt to create a duplicate workflow type
    res = client.post(
        "/api/workflow/add-type",
        json={
            "name": "complex_workflow_type",
            "description": "Duplicate workflow type",
            "model_requirements": [[{"component": "search", "type": "ndb"}]],
        },
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 400  # Expect failure due to duplicate name

    # Attempt to create workflow type with missing fields
    res = client.post(
        "/api/workflow/add-type",
        json={
            "name": "",
            "description": "",
            "model_requirements": [],
        },
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 422  # Validation error for missing fields

    # List workflow types to verify addition
    res = client.get("/api/workflow/types", headers=auth_header(admin_jwt))
    assert res.status_code == 200
    workflow_types = res.json()["data"]["types"]
    assert any(wt["name"] == "complex_workflow_type" for wt in workflow_types)

    # Attempt to delete the sample workflow type
    sample_workflow_type_id = next(
        wt["id"] for wt in workflow_types if wt["name"] == "sample_workflow_type"
    )
    # Access control check: Normal user trying to delete workflow type
    res = client.post(
        "/api/workflow/delete-type",
        json={"type_id": sample_workflow_type_id},
        headers=auth_header(normal_user_jwt),
    )
    assert res.status_code == 403  # Forbidden

    # global admin can delete the workflow
    res = client.post(
        "/api/workflow/delete-type",
        json={"type_id": sample_workflow_type_id},
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 200

    # Verify deletion of sample workflow type
    res = client.get("/api/workflow/types", headers=auth_header(admin_jwt))
    workflow_types_after_delete = res.json()["data"]["types"]
    assert not any(
        wt["name"] == "sample_workflow_type" for wt in workflow_types_after_delete
    )


### 2. Workflow Tests ###


def test_create_and_delete_workflow():
    from main import app

    client = TestClient(app)

    # Log in as workflow owner
    res = login(client, username="workflow-owner@mail.com", password="owner-password")
    assert res.status_code == 200
    owner_jwt = res.json()["data"]["access_token"]

    # Create a new workflow
    res = client.post(
        "/api/workflow/create",
        json={"name": "Test Workflow", "type_name": "complex_workflow_type"},
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 200
    workflow_id = res.json()["data"]["workflow_id"]

    # Verify that the workflow has been created
    res = client.get("/api/workflow/list", headers=auth_header(owner_jwt))
    assert res.status_code == 200
    workflows = res.json()["data"]
    assert any(wf["id"] == workflow_id for wf in workflows)

    # Access control check: Normal user trying to delete a workflow they do not own
    res = login(client, username="normal-user@mail.com", password="normal-password")
    normal_user_jwt = res.json()["data"]["access_token"]

    res = client.post(
        "/api/workflow/delete",
        json={"workflow_id": workflow_id},
        headers=auth_header(normal_user_jwt),
    )
    assert res.status_code == 403  # Forbidden

    # Global admin attempting to delete the workflow
    res = login(client, username="admin@mail.com", password="password")
    admin_jwt = res.json()["data"]["access_token"]

    res = client.post(
        "/api/workflow/delete",
        json={"workflow_id": workflow_id},
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 200  # Should succeed

    # Verify deletion by global admin
    res = client.get("/api/workflow/list", headers=auth_header(owner_jwt))
    assert res.status_code == 200
    workflows_after_delete = res.json()["data"]
    assert not any(wf["id"] == workflow_id for wf in workflows_after_delete)


### 3. Model Management in Workflows ###


def test_add_and_validate_models_to_workflow():
    from main import app

    client = TestClient(app)

    # Log in as workflow owner
    res = login(client, username="workflow-owner@mail.com", password="owner-password")
    assert res.status_code == 200
    owner_jwt = res.json()["data"]["access_token"]

    # Login in as normal user
    res = login(client, username="normal-user@mail.com", password="normal-password")
    assert res.status_code == 200
    normal_user_jwt = res.json()["data"]["access_token"]

    # Create a workflow for adding models
    res = client.post(
        "/api/workflow/create",
        json={"name": "Model Test Workflow", "type_name": "complex_workflow_type"},
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 200
    workflow_id = res.json()["data"]["workflow_id"]

    # Upload models required for workflow
    upload_model(client, owner_jwt, "model_1", "public")

    res = client.get("/api/model/public-list", params={"name": "model_1"})
    assert res.status_code == 200

    data = res.json()["data"]
    assert len(data) == 1

    model1_id = data[0]["model_id"]

    # Add models to the workflow
    res = client.post(
        "/api/workflow/add-models",
        json={
            "workflow_id": workflow_id,
            "model_ids": [model1_id],
            "components": ["search"],
        },
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 200

    # Validate the workflow
    res = client.post(
        "/api/workflow/validate",
        json={"workflow_id": workflow_id},
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 200
    assert "Validation successful" in res.json()["message"]

    # Access control check: Attempt to delete a model from workflow by non-owner
    res = client.post(
        "/api/workflow/delete-models",
        json={
            "workflow_id": workflow_id,
            "model_ids": [model1_id],
            "components": ["search"],
        },
        headers=auth_header(normal_user_jwt),
    )
    assert res.status_code == 403  # Forbidden

    # Delete model from workflow by owner
    res = client.post(
        "/api/workflow/delete-models",
        json={
            "workflow_id": workflow_id,
            "model_ids": [model1_id],
            "components": ["search"],
        },
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 200

    # Validate the workflow after deletion of a model
    res = client.post(
        "/api/workflow/validate",
        json={"workflow_id": workflow_id},
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 400  # Validation should fail due to missing models
