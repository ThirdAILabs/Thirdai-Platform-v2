import datetime
import os

import jwt


def create_token(expiration_min=15, **kwargs):
    """
    Creates a JWT token with the given expiration time and payload.

    Parameters:
    - expiration_min: int - The expiration time of the token in minutes (default: 15).
        Example: 15
    - **kwargs: Additional payload data to include in the token.
        Example: {"model_identifier": "user123/my_model", "user_id": "user123"}

    Returns:
    - str: The JWT token.
    """
    access_token_expires = datetime.datetime.utcnow() + datetime.timedelta(
        minutes=expiration_min
    )
    payload = {"exp": access_token_expires, **kwargs}

    access_token = jwt.encode(
        payload=payload, key=os.getenv("JWT_SECRET"), algorithm="HS256"
    )
    return access_token


def verify_token(token):
    """
    Verifies the given JWT token.

    Parameters:
    - token: str - The JWT token to verify.
        Example: "eyJhbGciOiJIUzI1NiIsInR5cCI..."

    Returns:
    - dict: The payload of the token if valid.

    Raises:
    - ValueError: If the token is not valid.
    """
    try:
        # This function automatically checks for token expiration:
        # https://pyjwt.readthedocs.io/en/stable/usage.html#expiration-time-claim-exp
        payload = jwt.decode(
            jwt=token, key=os.getenv("JWT_SECRET"), algorithms=["HS256"]
        )
        return payload
    except jwt.PyJWTError:
        raise ValueError("Token is not valid")
