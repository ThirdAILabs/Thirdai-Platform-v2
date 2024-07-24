import requests

BASE_URL = "http://localhost:8000"

def test_email_signup():
    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "test1user",
        "email": "test1user@mail.com",
        "password": "password123"
    })
    print("Test Email Signup: ", "Passed" if response.status_code == 200 else "Failed", response.json())

def test_email_signup_duplicate_email():
    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user2",
        "email": "user3@mail.com",
        "password": "password"
    })
    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user2",
        "email": "user3@mail.com",
        "password": "password"
    })
    print("Test Email Signup Duplicate Email: ", "Passed" if response.status_code == 400 else "Failed", response.json())

def test_email_signup_duplicate_username():
    requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user41",
        "email": "user4m@mail.com",
        "password": "password"
    })
    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "user41",
        "email": "user2m@mail.com",
        "password": "password"
    })
    print("Test Email Signup Duplicate Username: ", "Passed" if response.status_code == 400 else "Failed", response.json())

def test_add_admin():
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

    response = requests.post(f"{BASE_URL}/api/user/email-signup-basic", json={
        "username": "delete123",
        "email": "delete123@mail.com",
        "password": "password"
    })
    user_id = response.json()["data"]["user"]["user_id"]
    
    response = requests.get(f"{BASE_URL}/api/user/email-login", auth=("admin@mail.com", "password"))
    access_token = response.json()["data"]["access_token"]

    response = requests.delete(f"{BASE_URL}/api/user/delete-user", params={"user_id": user_id}, headers={"Authorization": f"Bearer {access_token}"})
    print("Test Delete User: ", "Passed" if response.status_code == 200 else "Failed", response.json())

if __name__ == "__main__":
    test_email_signup()
    test_email_signup_duplicate_email()
    test_email_signup_duplicate_username()
    test_add_admin()
    test_delete_user()
