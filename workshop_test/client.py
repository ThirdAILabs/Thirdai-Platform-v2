import requests
from requests.auth import HTTPBasicAuth
import json
import os
from typing import List, Dict, Optional
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

    def create_token_classifier(self, model_name: str, task_prompt: str, tags: List[dict], token: str, num_sentences: int = 10_000, num_samples_per_tag: Optional[int] = None) -> Dict:
        """
        Trains a token classification model that can detect the provided tags.
        It first generates training samples which contain tokens similar to the provided tag examples, then trains a model with this data.
        Parameters:
        - model_name: The name of the new model.
        - task_prompt: A prompt for the task that the model will perform, e.g. "Detect PII"
        - tags: A list of classes that tokens can be tagged with, accompanied by a description and examples, e.g.
            [
                {"name": "NAME", "examples": ["John Smith", "Anshumali Shrivastava"], "description": "A person's name"},
                {"name": "PHONE_NUMBER", "examples": ["123-123-1234", "123 123 1234", "(123)-123-1234"], "description": "American phone number"},
            ]
        - num_samples: The number of samples that will be generated for model training.
        - token: Authorization token from login.
        """
        if num_samples_per_tag is None:
            num_samples_per_tag = max((num_sentences // len(tags)), 50)
        # Set up the headers with authorization
        headers = {
            'Authorization': f'Bearer {token}'
        }
        # Prepare the form data
        form_data = {
            'datagen_options': json.dumps({
                'task_prompt': task_prompt,
                'datagen_options': {
                    'sub_type': 'token',
                    'task_prompt': task_prompt, # This can also be a distinct prompt about the domain of the task.
                    'tags': tags,
                    'num_sentences_to_generate': num_sentences,
                    'num_samples_per_tag': num_samples_per_tag,
                }
            })
        }
        # Make the POST request
        response = requests.post(
            f'{self.base_url}/api/train/nlp-datagen?model_name={model_name}',
            data=form_data,
            headers=headers
        )
        print(response.content)
        # Check for success
        response.raise_for_status()
        return response.json()

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

    def query_sentiment_model(self, model_id: str, query: str, token: str):
        """
        Retrieves top k most relevant references to the query from the deployed model.
        Parameters:
        - model_id: <username>/<modelname>
        - query: The query to search for.
        - token: Authorization token from login
        """
        # Define the URL for querying the deployed model
        headers = {"Authorization": f"Bearer {token}"}
        query_url = f"{self.base_url}/{model_id}/predict"
        # Set up the query parameters
        base_params = {"text": query, "top_k": 5}
        # Make a POST request to query the model
        response = requests.post(
            query_url,
            json=base_params,
            headers=headers,
        )
        # Check if the query was successful; if not, raise an exception
        if response.status_code != 200:
            raise Exception(f"Query failed: {response.status_code}, {response.text}")
        return response.json()["data"]["predicted_classes"]

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

    def token_classifier_predict(self, model_id: str, query: str, token: str, top_k=1):
        """
        Predicts the NER tags for a given query.
        Parameters:
        - model_id: model ID as returned by create_token_classifier. You can also find the model ID in the list returned by list_models.
        - query: The passage to predict the NER tags for.
        - token: Authorization token from login
        - top_k: The number of tags predicted for each token.
        Returns a dictionary in this format:
        {
            "text": "The text that was passed in",
            "predicted_tags": [
                ["TOP_TAG_FOR_FIRST_TOKEN", "SCORE_FOR_TOP_TAG_FOR_FIRST_TOKEN", ...],
                ["TOP_TAG_FOR_SECOND_TOKEN", "SCORE_FOR_TOP_TAG_FOR_SECOND_TOKEN", ...],
                ...
                ["TOP_TAG_FOR_LAST_TOKEN", "SCORE_FOR_TOP_TAG_FOR_LAST_TOKEN", ...]
            ]
        }
        The number of tags predicted for each token is specified in the top_k parameter.
        """
        headers = {"Authorization": f"Bearer {token}"}
        query_url = f"{self.base_url}/{model_id}/predict"
        base_params = {"text": query, "top_k": top_k}
        response = requests.post(
            query_url,
            json=base_params,
            headers=headers,
        )
        # Check if the query was successful; if not, raise an exception
        if response.status_code != 200:
            raise Exception(f"Query failed: {response.status_code}, {response.text}")
        return response.json()["data"]

    def obfuscate_pii(self, token_classifier_model_id: str, text_chunks: List[str], auth_token: str):
        """
        Obfuscates PII in the references using the NER model by replacing them with placeholders.
        Parameters:
        - token_classifier_model_id: model ID as returned by create_token_classifier. You can also find the model ID in the list returned by list_models.
        - text_chunks: A list of strings containing the text to obfuscate.
        - auth_token: Authorization token from login
        Returns a tuple containing:
        - A list of strings containing the obfuscated PII information.
        - A dictionary containing the mapping of obfuscated tokens to original tokens.
        """
        token_to_tag = {}
        token_counts = {}
        for text in text_chunks:
            text = " ".join(text.split())
            predicted_tags = self.token_classifier_predict(auth_token, token_classifier_model_id, text)
            predicted_tags = predicted_tags["predicted_tags"]
            for i, token in enumerate(text.split()):
                tag = predicted_tags[i][0]
                if tag != "O":
                    if token not in token_to_tag:
                        tg = f"<{tag}>"
                        token_to_tag[token] = tg
        token_counts = {v: 0 for k, v in token_to_tag.items()}
        inverse_map = {}
        for k, v in token_to_tag.items():
            new_tag = v[:-1] + f"_{token_counts[v]}>"
            inverse_map[new_tag] = k
            token_to_tag[k] = new_tag
            token_counts[v] += 1
        output_text = []
        for text in text_chunks:
            text = " ".join(text.split())
            redacted_text = [
                word if word not in token_to_tag else token_to_tag[word]
                for word in text.split()
            ]
            output_text.append(" ".join(redacted_text))
        return output_text, inverse_map

    def restore_pii(self, text: str, tag_to_token: Dict[str, str]):
        """
        Restores the PII in the text by replacing the placeholders with the original tokens.
        Parameters:
        - text: A string containing the obfuscated PII information.
        - tag_to_token: A dictionary containing the mapping of obfuscated tokens to original tokens.

        Returns a string containing the restored PII information.
        """
        restored_text = []
        for word in text.split():
            word = self.strip_non_alphanumeric(word)
            if word in tag_to_token.keys():
                restored_text.append(tag_to_token[word])
            else:
                restored_text.append(word)
        return " ".join(restored_text)


    def strip_non_alphanumeric(self, word):
        pattern = r"^[^a-zA-Z0-9_<>\s]+|[^a-zA-Z0-9_<>\s]+$"
        cleaned_string = re.sub(pattern, "", word)
        return cleaned_string

    def chat(self, model_id: str, user_input: str, token: str, session_id: str = None):
        """
        Sends a chat request to the /chat endpoint.
        Parameters:
        - user_input: The message or query from the user.
        - token: Authorization token from login.
        - session_id: (Optional) Session ID for maintaining conversation context.
        Returns:
        - Response from the chat API.
        """
        chat_url = f"{self.base_url}/{model_id}/chat"
        headers = {"Authorization": f"Bearer {token}"}
        payload = {
            "user_input": user_input,
            "session_id": session_id,
            "provider": "on-prem"
        }
        response = requests.post(chat_url, json=payload, headers=headers)
        if response.status_code != 200:
            raise Exception(f"Chat request failed: {response.status_code}, {response.text}")
        return response.json()