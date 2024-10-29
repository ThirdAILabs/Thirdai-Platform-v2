import traceback
from logging import Logger

from fastapi import Request
from platform_common.utils import response
from starlette import status
from starlette.middleware.base import BaseHTTPMiddleware


def create_log_request_response_middleware(logger: Logger):
    class LogRequestResponseMiddleware(BaseHTTPMiddleware):
        """
        Middleware to log request and response details for each endpoint.
        Logs request path, method, body, response status, and any errors.
        """

        def __init__(self, app):
            super().__init__(app)
            self.logger = logger

        async def dispatch(self, request: Request, call_next):
            # Log request information
            request_info = f"Request {request.method} {request.url}"

            try:
                # Call the next middleware or endpoint
                res = await call_next(request)

                res_body = b"".join([chunk async for chunk in res.body_iterator])
                self.logger.info(
                    f"{request_info} - Response Status: {res.status_code} - Response Body: {res_body.decode('utf-8')}"
                )

                # Reset the response body iterator for further processing
                async def body_iterator():
                    yield res_body

                res.body_iterator = body_iterator()
                return res

            except Exception as e:
                # Log errors with traceback and return custom error response
                error_trace = traceback.format_exc()
                self.logger.error(
                    f"Error during {request_info} - Error: {str(e)}\nTraceback:\n{error_trace}"
                )
                return response(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    message=str(traceback.format_exc()),
                    success=False,
                )

    return LogRequestResponseMiddleware
