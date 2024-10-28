import traceback
from logging import Logger

from fastapi import Request
from fastapi.responses import StreamingResponse
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
                # Extract and log request body (if any)
                body = await request.json()
                self.logger.info(f"{request_info} - Body: {body}")
            except Exception:
                self.logger.info(f"{request_info} - No JSON body")

            try:
                # Call the next middleware or endpoint
                response = await call_next(request)

                if not isinstance(response, StreamingResponse):
                    self.logger.info(
                        f"{request_info} - Response Status: {response.status_code} - Response Body: {response.body}"
                    )

                return response

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
