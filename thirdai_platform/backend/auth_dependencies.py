from fastapi import Depends, HTTPException, status
from sqlalchemy.orm import Session
from database.session import get_session
from auth.jwt import AuthenticatedUser, verify_access_token
from database import schema
from backend.utils import get_model_from_identifier, response


def verify_model_access(
    model_identifier: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):
    user = authenticated_user.user
    try:
        model: schema.Model = get_model_from_identifier(model_identifier, session)
    except Exception as error:
        return response(
            status_code=status.HTTP_400_BAD_REQUEST,
            message=str(error),
        )
    if model.user_id != user.id and not user.admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Access denied to the model"
        )


def get_current_user(
    session: Session = Depends(get_session),
    user: schema.User = Depends(verify_access_token),
):
    user = session.query(schema.User).filter(schema.User.id == user.user.id).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user"
        )
    return user


def verify_admin_access(current_user: schema.User = Depends(get_current_user)):
    if not current_user.admin:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, detail="Admin privileges required"
        )
    return current_user
