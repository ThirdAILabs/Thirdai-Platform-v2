# test_user_endpoints.py
import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from thirdai_platform.database import schema, Base
from thirdai_platform.database.session import get_session
from thirdai_platform.main import app

SQLALCHEMY_DATABASE_URL = "sqlite:///./test.db"
engine = create_engine(SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False})
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# dependency override to use the test database
def override_get_session():
    try:
        db = TestingSessionLocal()
        yield db
    finally:
        db.close()

app.dependency_overrides[get_session] = override_get_session

@pytest.fixture(scope="module")
def test_client():
    Base.metadata.create_all(bind=engine)
    with TestClient(app) as client:
        yield client
    Base.metadata.drop_all(bind=engine)

@pytest.fixture
def admin_token(test_client):
    response = test_client.post("/email-signup-basic", json={
        "username": "admin",
        "email": "admin@mail.com",
        "password": "password"
    })
    admin = response.json()["data"]["user"]

    db = next(override_get_session())
    user = db.query(schema.User).filter(schema.User.id == admin["user_id"]).first()
    user.admin = True
    db.commit()
    
    response = test_client.get("/email-login", auth=("admin@example.com", "adminpassword"))
    token = response.json()["data"]["access_token"]
    return token

@pytest.fixture
def normal_user_token(test_client):
    response = test_client.post("/email-signup-basic", json={
        "username": "normal",
        "email": "normal@mail.com",
        "password": "password"
    })
    normal_user = response.json()["data"]["user"]

    response = test_client.get("/email-login", auth=("user@example.com", "userpassword"))
    token = response.json()["data"]["access_token"]
    return token

def test_email_signup(test_client):
    response = test_client.post("/email-signup-basic", json={
        "username": "testuser",
        "email": "test@example.com",
        "password": "testpassword"
    })
    assert response.status_code == 200
    assert response.json()["message"] == "Successfully signed up via email."

def test_add_admin(test_client, admin_token):
    response = test_client.post("/add-admin", headers={"Authorization": f"Bearer {admin_token}"}, params={
        "email": "user@example.com"
    })
    assert response.status_code == 200
    assert response.json()["message"] == "User user@example.com has been successfully added as an admin."

def test_add_admin_not_allowed(test_client, normal_user_token):
    response = test_client.post("/add-admin", headers={"Authorization": f"Bearer {normal_user_token}"}, params={
        "email": "test@example.com"
    })
    assert response.status_code == 403
    assert response.json()["message"] == "You dont have enough permission to add another admin."

def test_delete_user(test_client, admin_token):
    response = test_client.post("/email-signup-basic", json={
        "username": "usertodelete",
        "email": "delete@example.com",
        "password": "deletepassword"
    })
    user_to_delete = response.json()["data"]["user"]

    response = test_client.delete(f"/delete-user/{user_to_delete['user_id']}", headers={"Authorization": f"Bearer {admin_token}"})
    assert response.status_code == 200
    assert response.json()["message"] == f"User with id {user_to_delete['user_id']} has been successfully deleted."

def test_delete_user_not_allowed(test_client, normal_user_token):
    response = test_client.delete(f"/delete-user/1", headers={"Authorization": f"Bearer {normal_user_token}"})
    assert response.status_code == 403
    assert response.json()["detail"] == "Admin privileges required"
