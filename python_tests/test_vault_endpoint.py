import unittest
import httpx

BASE_URL = "http://localhost:8000"


class TestVaultEndpoints(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        cls.client = httpx.Client(base_url=BASE_URL)

    def setUp(self):
        self.admin_credentials = {
            "username": "admin",
            "email": "admin@mail.com",
            "password": "password",
        }
        self.user_credentials = {
            "username": "test1user",
            "email": "test1user@mail.com",
            "password": "password123",
        }

        self.admin_headers = self.get_headers(self.admin_credentials)
        self.user_headers = self.get_headers(self.user_credentials)

    def get_headers(self, credentials):
        # Sign up the user
        signup_response = self.client.post(
            "/api/user/email-signup-basic",
            json={
                "username": credentials["username"],
                "email": credentials["email"],
                "password": credentials["password"],
            },
        )
        print(
            f"Signup Response for {credentials['username']}: {signup_response.json()}"
        )

        login_response = self.client.get(
            "/api/user/email-login",
            auth=(credentials["email"], credentials["password"]),
        )
        print(f"Login Response for {credentials['username']}: {login_response.json()}")

        self.assertEqual(login_response.status_code, 200)
        token = login_response.json()["data"]["access_token"]
        return {"Authorization": f"Bearer {token}"}

    def test_add_secret_admin(self):
        secret_data = {
            "user_id": "user123",
            "key": "AWS_ACCESS_TOKEN",
            "value": "aws_secret_value",
        }
        response = self.client.post(
            "/api/vault/add_secret", json=secret_data, headers=self.admin_headers
        )
        print(f"Add Secret Admin Response: {response.json()}")
        self.assertEqual(response.status_code, 200)
        self.assertDictEqual(response.json(), secret_data)

    def test_add_secret_user_forbidden(self):
        secret_data = {
            "user_id": "user123",
            "key": "AWS_ACCESS_TOKEN",
            "value": "aws_secret_value",
        }
        response = self.client.post(
            "/api/vault/add_secret", json=secret_data, headers=self.user_headers
        )
        print(f"Add Secret User Forbidden Response: {response.json()}")
        self.assertEqual(response.status_code, 403)
        self.assertEqual(response.json(), {"detail": "Admin privileges required"})

    def test_get_secret(self):
        secret_data = {
            "user_id": "user123",
            "key": "AWS_ACCESS_TOKEN",
            "value": "aws_secret_value",
        }
        add_response = self.client.post(
            "/api/vault/add_secret", json=secret_data, headers=self.admin_headers
        )
        print(f"Add Secret Response: {add_response.json()}")

        response = self.client.get(
            f"/api/vault/get_secret/{secret_data['user_id']}/{secret_data['key']}",
            headers=self.user_headers,
        )
        print(f"Get Secret Response: {response.json()}")
        self.assertEqual(response.status_code, 200)
        self.assertDictEqual(response.json(), secret_data)

    def test_get_secret_invalid_key(self):
        response = self.client.get(
            "/api/vault/get_secret/user123/INVALID_KEY", headers=self.user_headers
        )
        print(f"Get Secret Invalid Key Response: {response.json()}")
        self.assertEqual(response.status_code, 400)
        self.assertEqual(
            response.json(),
            {
                "detail": "Invalid key. Only 'AWS_ACCESS_TOKEN' and 'OPENAI_API_KEY' are allowed."
            },
        )

    def test_get_secret_not_found(self):
        response = self.client.get(
            "/api/vault/get_secret/user123/OPENAI_API_KEY", headers=self.user_headers
        )
        print(f"Get Secret Not Found Response: {response.json()}")
        self.assertEqual(response.status_code, 404)
        self.assertEqual(response.json(), {"detail": "Secret not found"})


if __name__ == "__main__":
    unittest.main()
