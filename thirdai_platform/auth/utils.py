import fastapi

# This is a helper object provided by FastAPI to extract the access token from a
# request. The paramter 'tokenURL' specifies which endpoint the user will use in
# the app to obtain the access token. When this is used in the fastapi.Depends
# clause in an endpoint (see `verify_access_token` below) it will find the
# 'Authorization' header in the request and check that the value of the header is
# 'Bearer <access token>' and then return the value of the access token within the
# header. It will return an 401 unauthorized code if the header is not found or
# is not properly formatted.
token_bearer = fastapi.security.OAuth2PasswordBearer(tokenUrl="/auth/token")

CREDENTIALS_EXCEPTION = fastapi.HTTPException(
    status_code=fastapi.status.HTTP_401_UNAUTHORIZED,
    detail="Invalid access token.",
    # This header indicates what type of authentication would be required to access
    # the resource.
    # https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
    headers={"WWW-Authenticate": "Bearer"},
)
