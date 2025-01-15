import os
import shutil

import pytest
import thirdai
import thirdai.neural_db
from fastapi.testclient import TestClient

from .utils import (
    add_user_to_team,
    assign_team_admin,
    auth_header,
    create_team,
    create_user,
    global_admin_token,
    login,
    upload_model,
)

pytestmark = [pytest.mark.unit]


def create_and_login(client, user):
    res = create_user(
        client, username=user, email=f"{user}@mail.com", password=f"{user}-pwd"
    )
    assert res.status_code == 200

    res = login(client, username=f"{user}@mail.com", password=f"{user}-pwd")
    assert res.status_code == 200

    return res.json()["data"]["access_token"]


@pytest.fixture(scope="session")
def create_models_and_users():
    from licensing.verify import verify_license
    from main import app

    client = TestClient(app)

    # So we can initialize NDBs vs upload
    verify_license.verify_and_activate(os.environ["LICENSE_PATH"])

    tokens = [create_and_login(client, user) for user in ["user_x", "user_y", "user_z"]]

    upload_model(client, tokens[0], "test_model_a", "public")
    upload_model(client, tokens[1], "test_model_b", "private")

    return client, tokens


def test_list_models(create_models_and_users):
    client, user_tokens = create_models_and_users

    for token, expected_models in [
        (user_tokens[0], ["user_x/test_model_a"]),
        (user_tokens[1], ["user_x/test_model_a", "user_y/test_model_b"]),
        (user_tokens[2], ["user_x/test_model_a"]),
    ]:
        res = client.get(
            "/api/model/list",
            headers=auth_header(token),
        )
        assert res.status_code == 200

        data = res.json()["data"]
        assert len(data) == len(expected_models)
        model_names = set(f"{m['username']}/{m['model_name']}" for m in data)
        assert model_names == set(expected_models)


def test_check_models(create_models_and_users):
    client, user_tokens = create_models_and_users

    res = client.get(
        "/api/model/name-check",
        params={"name": "test_model_a"},
        headers=auth_header(user_tokens[0]),
    )
    assert res.status_code == 200
    assert res.json()["data"]["model_present"]

    res = client.get(
        "/api/model/name-check",
        params={"name": "test_model_a"},
        headers=auth_header(user_tokens[1]),
    )
    assert res.status_code == 200
    assert not res.json()["data"]["model_present"]


def check_downloaded_model(res):
    extracted = ".test_data/download_model.ndb"

    with open(extracted + ".zip", "wb") as file:
        for chunk in res.iter_bytes(2000):
            file.write(chunk)

    shutil.unpack_archive(extracted + ".zip", extracted, format="zip")

    db = thirdai.neural_db.NeuralDB.from_checkpoint(extracted)
    assert db._savable_state.model.retriever.size() == 3

    shutil.rmtree(extracted)
    os.remove(extracted + ".zip")


def test_download_public_model(create_models_and_users):
    client, _ = create_models_and_users

    res = client.get(
        "/api/model/public-download", params={"model_identifier": "user_y/test_model_b"}
    )
    assert res.status_code == 403

    res = client.get(
        "/api/model/public-download", params={"model_identifier": "user_x/test_model_a"}
    )
    assert res.status_code == 200
    check_downloaded_model(res)
    res.close()


def test_download_model(create_models_and_users):
    client, user_tokens = create_models_and_users

    res = client.get(
        "/api/model/download",
        params={"model_identifier": "user_y/test_model_b"},
        headers=auth_header(user_tokens[0]),
    )
    assert res.status_code == 403

    res = client.get(
        "/api/model/download",
        params={"model_identifier": "user_y/test_model_b"},
        headers=auth_header(user_tokens[1]),
    )
    assert res.status_code == 200
    check_downloaded_model(res)
    res.close()


def test_list_all_models(create_models_and_users):
    client, _ = create_models_and_users

    global_admin = global_admin_token(client)

    res = client.get("/api/model/all-models", headers=auth_header(global_admin))
    assert res.status_code == 200

    assert len(res.json()["data"]) == 2


def test_update_access_level(create_models_and_users):
    client, user_tokens = create_models_and_users

    res = client.get(
        "/api/model/list",
        params={"name": "test_model", "access_level": ["public"]},
        headers=auth_header(user_tokens[0]),
    )
    assert res.status_code == 200
    assert ["test_model_a"] == [m["model_name"] for m in res.json()["data"]]

    res = client.post(
        "/api/model/update-access-level",
        params={"model_identifier": "user_x/test_model_a", "access_level": "private"},
        headers=auth_header(user_tokens[0]),
    )
    assert res.status_code == 200

    res = client.get(
        "/api/model/list",
        params={"name": "test_model", "access_level": ["public"]},
        headers=auth_header(user_tokens[0]),
    )
    assert res.status_code == 200
    assert [] == [m["model_name"] for m in res.json()["data"]]


def test_model_team_permissions(create_models_and_users):
    client, user_tokens = create_models_and_users

    global_admin = global_admin_token(client)
    res = create_team(client, "tmp_team", access_token=global_admin)
    assert res.status_code == 201
    team_id = res.json()["data"]["team_id"]

    for user in ["user_y@mail.com", "user_z@mail.com"]:
        res = add_user_to_team(
            client, team=team_id, user=user, access_token=global_admin
        )
        assert res.status_code == 200

    res = client.get(
        "/api/model/team-models",
        params={"team_id": team_id},
        headers=auth_header(global_admin),
    )
    assert res.status_code == 200
    assert len(res.json()["data"]) == 0

    # Only owner can add to team
    res = client.post(
        "/api/team/add-model-to-team",
        params={"model_identifier": "user_y/test_model_b", "team_id": team_id},
        headers=auth_header(user_tokens[2]),
    )
    assert res.status_code == 403

    # Owner can add model to team
    res = client.post(
        "/api/team/add-model-to-team",
        params={"model_identifier": "user_y/test_model_b", "team_id": team_id},
        headers=auth_header(user_tokens[1]),
    )
    assert res.status_code == 200

    res = client.get(
        "/api/model/team-models",
        params={"team_id": team_id},
        headers=auth_header(global_admin),
    )
    assert res.status_code == 200
    assert len(res.json()["data"]) == 1

    def check_model_permissions(owner, read, write):
        res = client.get(
            "/api/model/permissions",
            params={"model_identifier": "user_y/test_model_b"},
            headers=auth_header(user_tokens[1]),
        )
        assert res.status_code == 200
        permissions = res.json()["data"]

        assert set(owner) == set(p["username"] for p in permissions["owner"])
        assert set(read) == set(p["username"] for p in permissions["read"])
        assert set(write) == set(p["username"] for p in permissions["write"])

    check_model_permissions(
        owner=["admin", "user_y"],
        read=["user_z"],
        write=["user_y"],
    )

    res = client.post(
        "/api/model/update-default-permission",
        params={"model_identifier": "user_y/test_model_b", "new_permission": "write"},
        headers=auth_header(user_tokens[1]),
    )
    assert res.status_code == 200

    check_model_permissions(
        owner=["admin", "user_y"],
        read=[],
        write=["user_y", "user_z"],
    )

    res = client.post(
        "/api/model/update-default-permission",
        params={"model_identifier": "user_y/test_model_b", "new_permission": "read"},
        headers=auth_header(user_tokens[1]),
    )
    assert res.status_code == 200

    check_model_permissions(
        owner=["admin", "user_y"],
        read=["user_z"],
        write=["user_y"],
    )

    res = client.post(
        "/api/model/update-model-permission",
        params={
            "model_identifier": "user_y/test_model_b",
            "email": "user_z@mail.com",
            "permission": "write",
        },
        headers=auth_header(user_tokens[1]),
    )
    assert res.status_code == 200

    check_model_permissions(
        owner=["admin", "user_y"],
        read=[],
        write=["user_y", "user_z"],
    )


@pytest.fixture(scope="session")
def setup_users_and_models():
    from licensing.verify import verify_license
    from main import app

    client = TestClient(app)

    # License activation for ThirdAI NeuralDB
    verify_license.verify_and_activate(os.environ["LICENSE_PATH"])

    global_admin = global_admin_token(client)
    # creating team
    res = create_team(client, "team_1", access_token=global_admin)
    assert res.status_code == 201
    team_id = res.json()["data"]["team_id"]

    # Create and login users
    tokens = {
        "global_admin": global_admin,
        "user_1_model": create_and_login(client, "user_1_model"),
        "user_2_model": create_and_login(client, "user_2_model"),
        "user_3_model": create_and_login(client, "user_3_model"),
        "team_1": team_id,
    }
    # Assigning user_2_model and user_3_model to team_1
    res = add_user_to_team(
        client, team=team_id, user="user_2_model@mail.com", access_token=global_admin
    )
    assert res.status_code == 200

    res = assign_team_admin(client, team_id, "user_3_model@mail.com", global_admin)
    assert res.status_code == 200

    # Upload models
    upload_model(client, tokens["user_1_model"], "model5", "private")
    upload_model(client, tokens["user_2_model"], "model4", "private")
    upload_model(client, tokens["user_3_model"], "model3", "private")
    upload_model(client, tokens["user_3_model"], "model2", "private")
    upload_model(client, tokens["global_admin"], "model1", "public")

    res = client.post(
        "/api/model/update-access-level",
        params={
            "model_identifier": "user_3_model/model3",
            "access_level": "protected",
            "team_id": team_id,
        },
        headers=auth_header(tokens["user_3_model"]),
    )
    assert res.status_code == 200

    return client, tokens


def test_accessible_models_as_global_admin(setup_users_and_models):
    client, tokens = setup_users_and_models

    # Global admin should be able to access all models
    res = client.get("/api/model/list", headers=auth_header(tokens["global_admin"]))
    assert res.status_code == 200

    data = res.json()["data"]
    model_names = [model["model_name"] for model in data]
    assert set(model_names) == {
        "model1",
        "model2",
        "model3",
        "model4",
        "model5",
        "test_model_a",
        "test_model_b",
    }


def test_accessible_models_user_1_model(setup_users_and_models):
    client, tokens = setup_users_and_models

    # user_1_model should see only model1 (public) and model5 (private to them)
    res = client.get("/api/model/list", headers=auth_header(tokens["user_1_model"]))
    assert res.status_code == 200

    data = res.json()["data"]
    model_names = [model["model_name"] for model in data]
    assert set(model_names) == {"model1", "model5"}


def test_accessible_models_user_2_model(setup_users_and_models):
    client, tokens = setup_users_and_models

    # user_2_model should see model1 (public), model4 (private to them), and model3 (protected and owned by them)
    res = client.get("/api/model/list", headers=auth_header(tokens["user_2_model"]))
    assert res.status_code == 200

    data = res.json()["data"]
    model_names = [model["model_name"] for model in data]
    assert set(model_names) == {"model1", "model3", "model4"}


def test_accessible_models_user_3_model(setup_users_and_models):
    client, tokens = setup_users_and_models

    # user_3_model should see model1 (public), model2 (private to them), and model3 (protected and they belong to the same team)
    res = client.get("/api/model/list", headers=auth_header(tokens["user_3_model"]))
    assert res.status_code == 200

    data = res.json()["data"]
    model_names = [model["model_name"] for model in data]
    assert set(model_names) == {"model1", "model2", "model3"}


def test_accessible_models_no_access(setup_users_and_models):
    client, tokens = setup_users_and_models

    # user_3_model should not see model4 (private to user_2_model) or model5 (private to user_1_model)
    res = client.get("/api/model/list", headers=auth_header(tokens["user_3_model"]))
    assert res.status_code == 200

    data = res.json()["data"]
    model_names = [model["model_name"] for model in data]
    assert "model4" not in model_names
    assert "model5" not in model_names
