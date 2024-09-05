import pytest
from fastapi.testclient import TestClient

from .utils import auth_header, create_user, login, upload_model

pytestmark = [pytest.mark.unit]

### Helper Functions ###


def create_workflow(client, jwt_token, workflow_name, type_name):
    """
    Helper function to create a workflow.
    """
    res = client.post(
        "/api/workflow/create",
        params={"name": workflow_name, "type_name": type_name},
        headers=auth_header(jwt_token),
    )
    assert res.status_code == 200
    return res.json()["data"]["workflow_id"]


def delete_workflow(client, jwt_token, workflow_id, expected_status_code):
    """
    Helper function to delete a workflow.
    """
    res = client.post(
        "/api/workflow/delete",
        params={"workflow_id": workflow_id},
        headers=auth_header(jwt_token),
    )
    assert res.status_code == expected_status_code


def check_model_presence(client, jwt_token, model_name, should_exist=True):
    """
    Helper function to check if a model is present.
    """
    res = client.get(
        "/api/model/name-check",
        params={"name": model_name},
        headers=auth_header(jwt_token),
    )
    assert res.status_code == 200
    model_present = res.json()["data"]["model_present"]
    assert model_present == should_exist


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
            "name": "empty type",
            "description": "empty type",
            "model_requirements": [],
        },
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 400  # Cannot have empty model requirements.

    # Attempt to create workflow type with wrong parameters
    res = client.post(
        "/api/workflow/add-type",
        json={
            "name": "empty type",
            "description": "empty type",
            "model_requireme": {},
        },
        headers=auth_header(admin_jwt),
    )
    assert res.status_code == 422  # Cannot have empty model requiremnts.

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

    # Global admin can delete the workflow
    res = client.post(
        "/api/workflow/delete-type",
        params={"type_id": sample_workflow_type_id},
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
    workflow_id = create_workflow(
        client, owner_jwt, "Test Workflow", "complex_workflow_type"
    )

    # Verify that the workflow has been created
    res = client.get("/api/workflow/list", headers=auth_header(owner_jwt))
    assert res.status_code == 200
    workflows = res.json()["data"]
    assert any(wf["id"] == workflow_id for wf in workflows)

    # Access control check: Normal user trying to delete a workflow they do not own
    res = login(client, username="normal-user@mail.com", password="normal-password")
    assert res.status_code == 200
    normal_user_jwt = res.json()["data"]["access_token"]

    delete_workflow(
        client, normal_user_jwt, workflow_id, expected_status_code=403
    )  # Expect failure

    # Global admin attempting to delete the workflow
    res = login(client, username="admin@mail.com", password="password")
    assert res.status_code == 200
    admin_jwt = res.json()["data"]["access_token"]

    delete_workflow(
        client, admin_jwt, workflow_id, expected_status_code=200
    )  # Should succeed

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

    # Login as normal user
    res = login(client, username="normal-user@mail.com", password="normal-password")
    assert res.status_code == 200
    normal_user_jwt = res.json()["data"]["access_token"]

    # Create a workflow for adding models
    workflow_id = create_workflow(
        client, owner_jwt, "Model Test Workflow", "complex_workflow_type"
    )

    # Upload models required for workflow
    upload_model(client, owner_jwt, "model_1", "public")

    res = client.get("/api/model/public-list", params={"name": "model_1"})
    assert res.status_code == 200
    model1_id = res.json()["data"][0]["model_id"]

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
        params={"workflow_id": workflow_id},
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
        params={"workflow_id": workflow_id},
        headers=auth_header(owner_jwt),
    )
    assert res.status_code == 404  # Validation should fail due to missing models


def test_delete_workflow_delete_models():
    from main import app

    client = TestClient(app)

    # Log in as workflow owner
    res = login(client, username="workflow-owner@mail.com", password="owner-password")
    assert res.status_code == 200
    owner_jwt = res.json()["data"]["access_token"]

    # Create two workflows
    workflow_id_1 = create_workflow(
        client, owner_jwt, "Model Test Workflow 1", "complex_workflow_type"
    )
    workflow_id_2 = create_workflow(
        client, owner_jwt, "Model Test Workflow 2", "complex_workflow_type"
    )

    # Upload models required for workflow
    upload_model(client, owner_jwt, "model_a", "public")

    res = client.get("/api/model/public-list", params={"name": "model_a"})
    assert res.status_code == 200
    model_id = res.json()["data"][0]["model_id"]

    # Add model to both workflows
    for workflow_id in [workflow_id_1, workflow_id_2]:
        res = client.post(
            "/api/workflow/add-models",
            json={
                "workflow_id": workflow_id,
                "model_ids": [model_id],
                "components": ["search"],
            },
            headers=auth_header(owner_jwt),
        )
        assert res.status_code == 200

    # Delete workflow 2 and check model presence
    delete_workflow(client, owner_jwt, workflow_id_2, expected_status_code=200)
    check_model_presence(client, owner_jwt, "model_a", should_exist=True)

    # Delete workflow 1 and check model presence
    delete_workflow(client, owner_jwt, workflow_id_1, expected_status_code=200)
    check_model_presence(client, owner_jwt, "model_a", should_exist=False)
