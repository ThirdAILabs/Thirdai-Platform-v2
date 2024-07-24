import requests

BASE_URL = "http://localhost:8000"

def test_email_signup():
    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "testuser",
        "email": "testuser@mail.com",
        "password": "password123"
    })
    print("Test Email Signup: ", "Passed" if response.status_code == 200 else "Failed", response.json())

def test_email_signup_duplicate_email():
    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user1",
        "email": "user1@mail.com",
        "password": "password"
    })
    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user2",
        "email": "user1@mail.com",
        "password": "password"
    })
    print("Test Email Signup Duplicate Email: ", "Passed" if response.status_code == 400 else "Failed", response.json())

def test_email_signup_duplicate_username():
    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user1",
        "email": "user2@mail.com",
        "password": "password"
    })
    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user1",
        "email": "user3@mail.com",
        "password": "password"
    })
    print("Test Email Signup Duplicate Username: ", "Passed" if response.status_code == 400 else "Failed", response.json())

def test_add_admin():
    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "admin",
        "email": "admin@mail.com",
        "password": "password"
    })
    response = requests.get(f"{BASE_URL}/api/user/email-login", auth=("admin@mail.com", "password"))
    access_token = response.json()["data"]["access_token"]

    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user",
        "email": "user@mail.com",
        "password": "password"
    })
    response = requests.post(f"{BASE_URL}/api/user/add-admin", params={"email": "user@mail.com"}, headers={"Authorization": f"Bearer {access_token}"})
    print("Test Add Admin: ", "Passed" if response.status_code == 200 else "Failed", response.json())

def test_delete_user():
    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "admin",
        "email": "admin@mail.com",
        "password": "password"
    })
    response = requests.get(f"{BASE_URL}/api/user/email-login", auth=("admin@mail.com", "password"))
    access_token = response.json()["data"]["access_token"]

    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user",
        "email": "user@mail.com",
        "password": "password"
    })
    user_id = response.json()["data"]["user_id"]

    response = requests.delete(f"{BASE_URL}/api/user/delete-user/{user_id}", headers={"Authorization": f"Bearer {access_token}"})
    print("Test Delete User: ", "Passed" if response.status_code == 200 else "Failed", response.json())

if __name__ == "__main__":
    test_email_signup()
    test_email_signup_duplicate_email()
    test_email_signup_duplicate_username()
    test_add_admin()
    test_delete_user()
