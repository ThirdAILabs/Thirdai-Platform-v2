from fastapi import APIRouter, Depends, HTTPException, status
from database.session import get_session
from sqlalchemy.orm import Session
from database import schema
from backend.auth_dependencies import global_admin_only, team_admin_or_global_admin

team_router = APIRouter()


@team_router.post("/create-team", dependencies=[Depends(global_admin_only)])
def add_team(
    name: str,
    session: Session = Depends(get_session),
):
    team = session.query(schema.Team).filter(schema.Team.name == name).first()
    if team:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Team with this name already exists",
        )
    new_team = schema.Team(name=name)
    session.add(new_team)
    session.commit()
    session.refresh(new_team)
    return {"message": "Team created successfully", "team": new_team}


@team_router.post(
    "/add-user-to-team", dependencies=[Depends(team_admin_or_global_admin)]
)
def add_user_to_team(
    user_email: str,
    team_name: str,
    session: Session = Depends(get_session),
):
    user = session.query(schema.User).filter(schema.User.email == user_email).first()
    team = session.query(schema.Team).filter(schema.Team.name == team_name).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )
    user.team_id = team.id
    session.commit()
    return {"message": "User added to the team successfully"}


@team_router.post("/assign-team-admin", dependencies=[Depends(global_admin_only)])
def assign_team_admin(
    user_email: str,
    team_name: str,
    session: Session = Depends(get_session),
):
    user = session.query(schema.User).filter(schema.User.email == user_email).first()
    team = session.query(schema.Team).filter(schema.Team.name == team_name).first()
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )
    user.role = schema.Role.team_admin
    user.team_id = team.id
    session.commit()
    return {"message": "User assigned as team admin successfully"}


# Note(pratik): When a team is deleted, all the users and models in the team  are deleted too.
@team_router.delete("/delete-team", dependencies=[Depends(global_admin_only)])
def delete_team(
    team_name: str,
    session: Session = Depends(get_session),
):
    team = session.query(schema.Team).filter(schema.Team.name == team_name).first()
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )
    if session.query(schema.User).filter(schema.User.team_id == team.id).count() > 0:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Cannot delete a team with members",
        )
    session.delete(team)
    session.commit()
    return {"message": "Team deleted successfully"}
