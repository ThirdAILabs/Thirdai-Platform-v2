from backend.auth_dependencies import (
    global_admin_only,
    team_admin_or_global_admin,
    get_current_user,
)
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel
from sqlalchemy.orm import Session
from sqlalchemy.sql.expression import null
from backend.utils import response

team_router = APIRouter()


class CreateTeamInput(BaseModel):
    name: str


class AddModelInput(BaseModel):
    model_name: str
    team_name: str


class RemoveModelInput(BaseModel):
    model_name: str


class AddUserToTeamInput(BaseModel):
    user_email: str
    team_name: str
    role: schema.Role = (
        schema.Role.user
    )  # default role is user to a team if not specified


class AssignTeamAdminInput(BaseModel):
    user_email: str
    team_name: str


class DeleteTeamInput(BaseModel):
    team_name: str


@team_router.post("/create-team", dependencies=[Depends(global_admin_only)])
def add_team(input: CreateTeamInput, session: Session = Depends(get_session)):
    if session.query(schema.Team).filter_by(name=input.name).first():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Team with this name already exists",
        )
    new_team = schema.Team(name=input.name)
    session.add(new_team)
    session.commit()
    session.refresh(new_team)
    return response(
        status_code=status.HTTP_201_CREATED,
        message="Team created successfully",
        data={"team_id": str(new_team.id), "team_name": new_team.name},
    )


@team_router.post(
    "/add-user-to-team", dependencies=[Depends(team_admin_or_global_admin)]
)
def add_user_to_team(
    input: AddUserToTeamInput,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    user = session.query(schema.User).filter_by(email=input.user_email).first()
    team = session.query(schema.Team).filter_by(name=input.team_name).first()

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    # check if the current user is a team admin of the same team
    if not current_user.is_team_admin_of_team(team.id):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User does not have permission to add members to this team",
        )

    user_team = (
        session.query(schema.UserTeam)
        .filter_by(user_id=user.id, team_id=team.id)
        .first()
    )
    if user_team:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="User already a member of this team",
        )

    new_user_team = schema.UserTeam(user_id=user.id, team_id=team.id, role=input.role)
    session.add(new_user_team)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="User added to the team successfully",
        data={"user_id": str(user.id), "team_id": str(team.id)},
    )


@team_router.post(
    "/assign-team-admin", dependencies=[Depends(team_admin_or_global_admin)]
)
def assign_team_admin(
    input: AssignTeamAdminInput,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
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

    # check if the current user is a team admin of the same team
    if not current_user.is_team_admin_of_team(team.id):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User does not have permission to assign team admin for this team",
        )

    user_team = (
        session.query(schema.UserTeam)
        .filter_by(user_id=user.id, team_id=team.id)
        .first()
    )

    if user_team:
        user_team.role = schema.Role.team_admin
    else:
        new_user_team = schema.UserTeam(
            user_id=user.id, team_id=team.id, role=schema.Role.team_admin
        )
        session.add(new_user_team)

    session.commit()
    return response(
        status_code=status.HTTP_200_OK,
        message=f"User {input.user_email} has been successfully assigned as team admin for {input.team_name}.",
        data={"user_id": str(user.id), "team_id": str(team.id)},
    )


@team_router.delete("/delete-team", dependencies=[Depends(global_admin_only)])
def delete_team(input: DeleteTeamInput, session: Session = Depends(get_session)):
    team = session.query(schema.Team).filter_by(name=input.team_name).first()
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    # remove all users from the team
    user_teams = session.query(schema.UserTeam).filter_by(team_id=team.id).all()
    for user_team in user_teams:
        session.delete(user_team)

    # setting the team_id of all models belonging to this team to NULL
    models = session.query(schema.Model).filter_by(team_id=team.id).all()
    for model in models:
        # since one model belongs to only one team
        model.access_level = schema.Access.private

        model.team_id = null()

    session.commit()

    # deleting the team
    session.delete(team)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Team deleted successfully",
        data={"team_id": str(team.id)},
    )


@team_router.post("/add-model-to-team")
def add_model_to_team(
    input: AddModelInput,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    model = session.query(schema.Model).filter_by(name=input.model_name).first()
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Model not found"
        )
    if model.team_id is not None:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Model already belongs to a team",
        )

    team = session.query(schema.Team).filter_by(name=input.team_name).first()
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    if model.get_user_permission(current_user) != schema.Permission.write:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User does not have enough access to the model.",
        )

    model.team_id = team.id
    if model.access == schema.Access.private:
        model.access = schema.Access.protected

    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Model added to team successfully",
        data={"model_id": str(model.id), "team_id": str(team.id)},
    )


@team_router.post("/remove-model-from-team")
def remove_model_from_team(
    input: RemoveModelInput,
    session: Session = Depends(get_session),
    current_user: schema.User = Depends(get_current_user),
):
    model = session.query(schema.Model).filter_by(name=input.model_name).first()
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Model not found"
        )

    # check if the current user belongs to the team the model is associated with
    if model.team_id:
        if model.get_user_permission(current_user) != schema.Permission.write:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="User does not have enough access to the model.",
            )

    model.team_id = null()
    model.access = schema.Access.private

    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Model removed from team successfully",
        data={"model_id": str(model.id)},
    )
