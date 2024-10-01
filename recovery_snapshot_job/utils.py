import logging
from functools import wraps

from fastapi import HTTPException, status


def handle_exceptions(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            class_name = args[0].__class__.__name__ if args else "UnknownClass"
            method_name = func.__name__
            logging.error(
                f"Error in class '{class_name}', method '{method_name}' "
                f"with arguments {args[1:]}, and keyword arguments {kwargs}. "
                f"Error: {str(e)}"
            )
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"An error occurred: {str(e)}",
            )

    return wrapper
