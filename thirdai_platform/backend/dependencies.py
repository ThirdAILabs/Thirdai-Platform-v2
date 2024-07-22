from fastapi import Depends, HTTPException, status
from sqlalchemy.orm import Session
from database import get_session
from auth.jwt import AuthenticatedUser, create_access_token, verify_access_token
from database import schema

def get_model_from_identifier(model_identifier: str, session: Session):
    model = session.query(schema.Model).filter_by(identifier=model_identifier).first()
    if not model:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Model not found")
    return model

def verify_model_access(
    model_identifier: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token)
):
    user = authenticated_user.user
    model = get_model_from_identifier(model_identifier, session)
    if model.user_id != user.id:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Access denied to the model")
    return model

def get_current_user(session: Session = Depends(get_session), token: str = Depends(verify_access_token)):
    user_id = verify_access_token(token)
    user = session.query(schema.User).filter(schema.User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user")
    return user

def verify_admin_access(current_user: schema.User = Depends(get_current_user)):
    if not current_user.admin:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Admin privileges required")
    return current_user