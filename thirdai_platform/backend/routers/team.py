from backend.auth_dependencies import (
    global_admin_only,
    is_model_owner,
    team_admin_or_global_admin,
)
from backend.utils import get_model_from_identifier, response
from database import schema
from database.session import get_session
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

team_router = APIRouter()


@team_router.post("/create-team", dependencies=[Depends(global_admin_only)])
def add_team(name: str, session: Session = Depends(get_session)):
    if session.query(schema.Team).filter_by(name=name).first():
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Team with this name already exists",
        )
    new_team = schema.Team(name=name)
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
    email: str,
    team_id: str,
    role: schema.Role = schema.Role.user,
    session: Session = Depends(get_session),
):
    user = session.query(schema.User).filter_by(email=email).first()
    team = session.query(schema.Team).filter_by(id=team_id).first()

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
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

    new_user_team = schema.UserTeam(user_id=user.id, team_id=team.id, role=role)
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
    email: str,
    team_id: str,
    session: Session = Depends(get_session),
):
    user: schema.User = (
        session.query(schema.User).filter(schema.User.email == email).first()
    )
    team: schema.Team = session.query(schema.Team).get(team_id)

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User not found, {email}",
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Team not found, {team_id}",
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
        message=f"User {email} has been successfully assigned as team admin for {team.name}.",
        data={"user_id": str(user.id), "team_id": str(team.id)},
    )


@team_router.delete("/delete-team", dependencies=[Depends(global_admin_only)])
def delete_team(team_id: str, session: Session = Depends(get_session)):
    team: schema.Team = session.query(schema.Team).get(team_id)
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

        model.team_id = None

    session.commit()

    # deleting the team
    session.delete(team)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Team deleted successfully",
        data={"team_id": str(team.id)},
    )


@team_router.post("/add-model-to-team", dependencies=[Depends(is_model_owner)])
def add_model_to_team(
    model_identifier: str,
    team_id: str,
    session: Session = Depends(get_session),
):
    model = get_model_from_identifier(model_identifier, session)
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Model not found"
        )
    if model.team_id is not None:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Model already belongs to a team",
        )

    team: schema.Team = session.query(schema.Team).get(team_id)
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    model.team_id = team.id
    model.access_level = schema.Access.protected

    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Model added to team successfully",
        data={"model_id": str(model.id), "team_id": str(team.id)},
    )


@team_router.post("/remove-model-from-team", dependencies=[Depends(is_model_owner)])
def remove_model_from_team(
    model_identifier: str,
    session: Session = Depends(get_session),
):
    model = get_model_from_identifier(model_identifier, session)
    if not model:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Model not found"
        )

    model.team_id = None
    model.access_level = schema.Access.private

    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Model removed from team successfully",
        data={"model_id": str(model.id)},
    )


@team_router.post(
    "/remove-user-from-team", dependencies=[Depends(team_admin_or_global_admin)]
)
def remove_user_from_team(
    email: str,
    team_id: str,
    session: Session = Depends(get_session),
):
    user: schema.User = session.query(schema.User).filter_by(email=email).first()
    team: schema.Team = session.query(schema.Team).get(team_id)

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    user_team = (
        session.query(schema.UserTeam)
        .filter_by(user_id=user.id, team_id=team.id)
        .first()
    )
    if not user_team:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="User is not a member of this team",
        )

    models = (
        session.query(schema.Model).filter_by(user_id=user.id, team_id=team.id).all()
    )
    for model in models:
        model.access_level = schema.Access.private
        model.team_id = None

    session.delete(user_team)
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="User removed from the team successfully",
        data={"user_id": str(user.id), "team_id": str(team.id)},
    )


@team_router.post(
    "/remove-team-admin", dependencies=[Depends(team_admin_or_global_admin)]
)
def remove_team_admin(
    email: str,
    team_id: str,
    session: Session = Depends(get_session),
):
    user: schema.User = session.query(schema.User).filter_by(email=email).first()
    team: schema.Team = session.query(schema.Team).get(team_id)

    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="User not found"
        )
    if not team:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Team not found"
        )

    user_team = (
        session.query(schema.UserTeam)
        .filter_by(user_id=user.id, team_id=team.id)
        .first()
    )
    if not user_team or user_team.role != schema.Role.team_admin:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="User is not a team admin of this team",
        )

    user_team.role = schema.Role.user
    session.commit()

    return response(
        status_code=status.HTTP_200_OK,
        message="Team admin role removed successfully",
        data={"user_id": str(user.id), "team_id": str(team.id)},
    )
