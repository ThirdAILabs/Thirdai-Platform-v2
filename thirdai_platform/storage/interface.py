class StorageInterface:
    def create_upload_token(self, model_identifier, user_id, model_id, expiration_min):
        """
        Creates an upload token for a model.

        Parameters:
        - model_identifier: str - The unique identifier for the model.
            Example: "user123/my_model"
        - user_id: str - The ID of the user uploading the model.
            Example: "user123"
        - model_id: str - The ID of the model to be uploaded.
            Example: "model456"
        - expiration_min: int - The expiration time of the token in minutes.
            Example: 15

        Returns:
        - str: The upload token.
        """
        raise NotImplementedError

    def verify_upload_token(self, token):
        """
        Verifies the given upload token.

        Parameters:
        - token: str - The upload token to verify.
            Example: "eyJhbGciOiJIUzI1NiIsInR5cCI..."

        Returns:
        - dict: The payload of the token if valid.
        """
        raise NotImplementedError

    def upload_chunk(
        self,
        model_id: str,
        chunk_data: bytes,
        chunk_number: int,
        model_type: str,
        compressed: bool,
    ):
        """
        Uploads a chunk of the model.

        Parameters:
        - model_id: str - The ID of the model.
            Example: "model456"
        - chunk_data: bytes - The raw bytes of the chunk.
        - chunk_number: int - The position of the chunk.
            Example: 1
        - model_type: str - The type of the model (e.g., "ndb").
            Example: "ndb"
        - compressed: bool - Whether the chunk is compressed.
            Example: True
        """
        raise NotImplementedError

    def commit_upload(
        self, model_id: str, total_chunks: int, model_type: str, compressed: bool
    ):
        """
        Commits the upload after all chunks have been uploaded.

        Parameters:
        - model_id: str - The ID of the model.
            Example: "model456"
        - total_chunks: int - The total number of chunks uploaded.
            Example: 10
        - model_type: str - The type of the model (e.g., "ndb").
            Example: "ndb"
        - compressed: bool - Whether the model is compressed.
            Example: True
        """
        raise NotImplementedError

    def prepare_download(self, model_id: str, model_type: str, compressed: bool):
        """
        Prepares the model for download.

        Parameters:
        - model_id: str - The ID of the model.
            Example: "model456"
        - model_type: str - The type of the model (e.g., "ndb").
            Example: "ndb"
        - compressed: bool - Whether the model is compressed.
            Example: True
        """
        raise NotImplementedError

    def download_chunk_stream(
        self, model_id: str, block_size: int, model_type: str, compressed: bool
    ):
        """
        Streams the download of the model in chunks.

        Parameters:
        - model_id: str - The ID of the model.
            Example: "model456"
        - block_size: int - The size of each chunk in bytes.
            Example: 8192
        - model_type: str - The type of the model (e.g., "ndb").
            Example: "ndb"
        - compressed: bool - Whether the model is compressed.
            Example: True

        Returns:
        - generator: A generator that yields chunks of the model.
        """
        raise NotImplementedError

    def delete(self, model_id: str):
        """
        Deletes the model.

        Parameters:
        - model_id: str - The ID of the model.
            Example: "model456"
        """
        raise NotImplementedError
