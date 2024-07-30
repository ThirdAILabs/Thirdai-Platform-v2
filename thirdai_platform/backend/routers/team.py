from backend.auth_dependencies import global_admin_only, team_admin_or_global_admin
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel
from sqlalchemy.orm import Session

team_router = APIRouter()


class CreateTeamInput(BaseModel):
    name: str


class AddUserToTeamInput(BaseModel):
    user_email: str
    team_name: str


class AssignTeamAdminInput(BaseModel):
    user_email: str
    team_name: str


class DeleteTeamInput(BaseModel):
    team_name: str


@team_router.post("/create-team", dependencies=[Depends(global_admin_only)])
def add_team(
    input: CreateTeamInput,
    session: Session = Depends(get_session),
):
    team = session.query(schema.Team).filter(schema.Team.name == input.name).first()
    if team:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Team with this name already exists",
        )
    new_team = schema.Team(name=input.name)
    session.add(new_team)
    session.commit()
    session.refresh(new_team)
    return {
        "status": "success",
        "message": "Team created successfully",
        "team": new_team,
    }


@team_router.post(
    "/add-user-to-team", dependencies=[Depends(team_admin_or_global_admin)]
)
def add_user_to_team(
    input: AddUserToTeamInput,
    session: Session = Depends(get_session),
):
    user = (
        session.query(schema.User).filter(schema.User.email == input.user_email).first()
    )
    team = (
        session.query(schema.Team).filter(schema.Team.name == input.team_name).first()
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User not found, {input.user_email}",
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Team not found, {input.team_name}",
        )
    user.team_id = team.id
    session.commit()
    return {"status": "success", "message": "User added to the team successfully"}


@team_router.post("/assign-team-admin", dependencies=[Depends(global_admin_only)])
def assign_team_admin(
    input: AssignTeamAdminInput,
    session: Session = Depends(get_session),
):
    user = (
        session.query(schema.User).filter(schema.User.email == input.user_email).first()
    )
    team = (
        session.query(schema.Team).filter(schema.Team.name == input.team_name).first()
    )
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User not found, {input.user_email}",
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Team not found, {input.team_name}",
        )
    user.role = schema.Role.team_admin
    user.team_id = team.id
    session.commit()
    return {"status": "success", "message": "User assigned as team admin successfully"}


@team_router.delete("/delete-team", dependencies=[Depends(global_admin_only)])
def delete_team(
    input: DeleteTeamInput,
    session: Session = Depends(get_session),
):
    team = (
        session.query(schema.Team).filter(schema.Team.name == input.team_name).first()
    )
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
    return {"status": "success", "message": "Team deleted successfully"}
