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
