import requests
from requests.auth import HTTPBasicAuth
import json
import os
from typing import List, Optional
import time
import re

# base_url = "http://34.236.171.80"


class PlatformClient:
    def __init__(self, base_url):
        self.base_url = base_url

    def signup(self, username: str, email: str, password: str):
        """
        Signup a user with the given username, email, and password.
        Once the request is sent, the user will receive an email with a link to verify their account.
        """
        url = f"{self.base_url}/api/user/email-signup-basic"
        payload = {"username": username, "email": email, "password": password}
        response = requests.post(url, json=payload)
        return response.json()

    def login(self, email: str, password: str):
        """
        Login a user with the given email and password.
        If successful, this function returns an access token.
        """
        url = f"{self.base_url}/api/user/email-login"
        response = requests.get(url, auth=HTTPBasicAuth(email, password))
        result = response.json()
        token = result.get("data", {}).get("access_token")
        return token

    def list_models(
        self, name: str, token: str, domain=None, username=None, type=None, sub_type=None, access_level=None
    ):
        """
        List models based on filters for authenticated users.
        Returns:
        - List of models that match the provided filters.
        """
        url = f"{self.base_url}/api/model/list"
        headers = {"Authorization": f"Bearer {token}"}
        params = {
            "name": name,
            "domain": domain,
            "username": username,
            "type": type,
            "sub_type": sub_type,
            "access_level": access_level,
        }
        response = requests.get(
            url, headers=headers, params={k: v for k, v in params.items() if v is not None}
        )
        return response.json()

    def delete_model(self, model_identifier: str, token: str):
        """
        Delete a specified model. model_identifier is username/modelname
        """
        url = f"{self.base_url}/api/model/delete"
        headers = {"Authorization": f"Bearer {token}"}
        params = {"model_identifier": model_identifier}
        response = requests.post(url, headers=headers, params=params)
        return response.json()

    def create_retriever_model(
        self, 
        model_name: str,
        token: str,
        base_model_identifier: Optional[str] = None,
        files: Optional[List[str]] = None,  # Local file paths
        s3_urls: Optional[List[str]] = None,  # S3 URLs
        nfs_paths: Optional[List[str]] = None,  # NFS paths
    ):
        """
        Creates and trains an NDB retriever model with local, S3, and NFS files.
        Parameters:
        - model_name: The name of the model.
        - token: Authorization token from login.
        - base_model_identifier: (Optional) The identifier of the base model to use.
        - files: (Optional) List of local file paths.
        - s3_urls: (Optional) List of S3 URLs for files.
        - nfs_paths: (Optional) List of NFS paths for files.
        """
        url = f"{self.base_url}/api/train/ndb"
        headers = {"Authorization": f"Bearer {token}"}
        # Construct file information based on the different sources
        file_info = {"unsupervised_files": [], "supervised_files": [], "test_files": []}
        if files:
            for file_path in files:
                file_info["unsupervised_files"].append(
                    {"path": file_path, "location": "local"}
                )
        if s3_urls:
            for s3_url in s3_urls:
                file_info["unsupervised_files"].append({"path": s3_url, "location": "s3"})
        if nfs_paths:
            for nfs_path in nfs_paths:
                file_info["unsupervised_files"].append(
                    {"path": nfs_path, "location": "nfs"}
                )
        # Prepare the files for upload
        upload_files = []
        if files:
            for file_path in files:
                if os.path.isfile(file_path):
                    upload_files.append(("files", open(file_path, "rb")))
        print(upload_files)
        upload_files.append(
            ("file_info", (None, json.dumps(file_info), "application/json"))
        )
        print(upload_files)
        try:
            # Send the POST request to the /ndb endpoint
            response = requests.post(
                url,
                headers=headers,
                params={
                    "model_name": model_name,
                    "base_model_identifier": base_model_identifier,
                },
                files=upload_files,
            )
            # Check for the response
            if response.status_code == 200:
                print("Model training job submitted successfully.")
                print(response.json())
                return response.json()["data"]["model_id"]
            else:
                print("Failed to submit the model training job.")
                print(response.json())
        except Exception as e:
            print(f"An error occurred: {str(e)}")

    def await_train(self, model_identifier: str, token: str):
        """
        Blocks until the model has finished training.
        Parameters:
        - model_identifier: <username>/<modelname>
        - token: Authorization token from login
        """
        # Define the URL for checking training status
        status_url = f"{self.base_url}/api/train/status"
        headers = {"Authorization": f"Bearer {token}"}
        while True:
            # Make a GET request to check the training status
            response = requests.get(
                status_url, params={"model_identifier": model_identifier}, headers=headers
            )
            if response.status_code != 200:
                raise Exception(
                    f"Failed to get training status: {response.status_code}, {response.text}"
                )
            # Check the training status
            status = response.json()["data"]["train_status"]
            if status == "complete":
                print("Training completed successfully.")
                break
            elif status == "failed":
                raise Exception("Training failed.")
            print("Training in progress...")
            time.sleep(10)  # Wait for 10 seconds before checking again

    def insert_retrieval_model_documents(self, model_id: str, local_files: List[str], token: str):
        """
        Inserts documents into an existing retrieval model.
        Parameters:
        - model_id: model ID as returned by create_retrieval_model. You can also find the model ID in the list returned by list_models.
        - local_files: List of local file paths.
        - token: Authorization token from login
        """
        headers = {"Authorization": f"Bearer {token}"}
        query_url = f"{self.base_url}/{model_id}/insert"
        files = [("files", open(local_file, "rb")) for local_file in local_files]
        documents = [self.create_doc_dict(local_file, "local") for local_file in local_files]
        files.append(("documents", (None, json.dumps(documents), "application/json")))
        response = requests.post(
            query_url,
            files=files,
            headers=headers,
        )
        return response.json()


    def create_doc_dict(self, path: str, doc_type: str):
        """
        Creates a document dictionary for different document types.
        Parameters:
        path (str): Path to the document file.
        doc_type (str): Type of the document location.
        Returns:
        dict[str, str]: Dictionary containing document details.
        Raises:
        Exception: If the document type is not supported.
        """
        _, ext = os.path.splitext(path)
        if ext == ".pdf":
            return {"document_type": "PDF", "path": path, "location": doc_type}
        if ext == ".csv":
            return {"document_type": "CSV", "path": path, "location": doc_type}
        if ext == ".docx":
            return {"document_type": "DOCX", "path": path, "location": doc_type}
        raise Exception(f"Please add a map from {ext} to document dictionary.")

    def await_train(self, model_identifier: str, token: str):
        """
        Blocks until the model has finished training.
        Parameters:
        - model_identifier: <username>/<modelname>
        - token: Authorization token from login
        """
        # Define the URL for checking training status
        status_url = f"{self.base_url}/api/train/status"
        headers = {"Authorization": f"Bearer {token}"}
        while True:
            # Make a GET request to check the training status
            response = requests.get(
                status_url, params={"model_identifier": model_identifier}, headers=headers
            )
            if response.status_code != 200:
                raise Exception(
                    f"Failed to get training status: {response.status_code}, {response.text}"
                )
            # Check the training status
            status = response.json()["data"]["train_status"]
            if status == "complete":
                print("Training completed successfully.")
                break
            elif status == "failed":
                raise Exception("Training failed.")
            print("Training in progress...")
            time.sleep(10)  # Wait for 10 seconds before checking again

    def deploy_model(self, model_identifier: str, token: str):
        """
        Allocates resources to serve the model.
        Parameters:
        - model_identifier: <username>/<modelname>
        - token: Authorization token from login
        """
        # Define the URL for model deployment
        deploy_url = f"{self.base_url}/api/deploy/run"
        headers = {"Authorization": f"Bearer {token}"}
        params = {"model_identifier": model_identifier}
        # Make a POST request to deploy the model
        response = requests.post(deploy_url, headers=headers, params=params)
        # Extract deployment ID from the response
        content = response.json()
        deployment_id = content["data"]["model_id"]
        print(f"Model deployed successfully. Deployment ID: {deployment_id}")
        return deployment_id

    def await_deploy(self, model_identifier: str, token: str):
        """
        Blocks until the model is deployed.
        Parameters:
        - model_identifier: <username>/<modelname>
        - token: Authorization token from login
        """
        # Define the URL for checking deployment status
        status_url = f"{self.base_url}/api/deploy/status"
        headers = {"Authorization": f"Bearer {token}"}
        while True:
            # Make a GET request to check the deployment status
            response = requests.get(
                status_url, params={"model_identifier": model_identifier}, headers=headers
            )
            if response.status_code != 200:
                raise Exception(
                    f"Failed to get deployment status: {response.status_code}, {response.text}"
                )
            # Check the deployment status
            status = response.json()["data"]["deploy_status"]
            if status == "complete":
                print("Deployment completed successfully.")
                break
            elif status == "failed":
                raise Exception("Deployment failed.")
            print("Deployment in progress...")
            time.sleep(10)  # Wait for 10 seconds before checking again

    def query_retrieval_model(self, model_id: str, query: str, token: str):
        """
        Retrieves top k most relevant references to the query from the deployed model.
        Parameters:
        - model_id: <username>/<modelname>
        - query: The query to search for.
        - token: Authorization token from login
        """
        # Define the URL for querying the deployed model
        headers = {"Authorization": f"Bearer {token}"}
        query_url = f"{self.base_url}/{model_id}/search"
        # Set up the query parameters
        base_params = {"query": query, "top_k": 5}
        # Make a POST request to query the model
        response = requests.post(
            query_url,
            json=base_params,
            headers=headers,
        )
        # Check if the query was successful; if not, raise an exception
        if response.status_code != 200:
            raise Exception(f"Query failed: {response.status_code}, {response.text}")
        return response.json()["data"]["references"]

    def upvote_reference(self, model_id: str, query: str, reference_id: str, token: str):
        """
        Upvotes a reference for a given query.
        Parameters:
        - model_id: model ID as returned by create_retrieval_model. You can also find the model ID in the list returned by list_models.
        - query: The query for which the reference is upvoted.
        - reference_id: The ID of the reference to upvote.
        - token: Authorization token from login
        """
        headers = {"Authorization": f"Bearer {token}"}
        query_url = f"{self.base_url}/{model_id}/upvote"
        # Set up the query parameters
        text_id_pairs = [{"query_text": query, "reference_id": reference_id}]
        # Make a POST request to query the model
        response = requests.post(
            query_url,
            json={"text_id_pairs": text_id_pairs},
            headers=headers,
        )
        return response.json()
