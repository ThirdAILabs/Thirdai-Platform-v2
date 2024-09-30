import os
from urllib.parse import urlencode, urljoin

from backend.mailer import Mailer
from database import schema

from typing import List


def send_verification_mail(self, email: str, verification_token: str, username: str):
    """
    Send verification email to the user_to_delete.
    """
    subject = "Verify Your Email Address"
    base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
    args = {"verification_token": verification_token}
    verify_link = urljoin(base_url, f"api/user/redirect-verify?{urlencode(args)}")
    body = f"<p>Please click the following link to verify your email address: <a href='{verify_link}'>verify</a></p>"

    Mailer(to=f"{username} <{email}>", subject=subject, body=body)


def delete_all_models_for_user(user_to_delete, session):
    team_admins: List[schema.UserTeam] = (
        session.query(schema.UserTeam).filter_by(role=schema.Role.team_admin).all()
    )
    team_admin_map = {
        team_admin.team_id: team_admin.user_id for team_admin in team_admins
    }

    models: List[schema.Model] = user_to_delete.models

    for model in models:
        if model.access_level == schema.Access.protected:
            new_owner_id = team_admin_map.get(model.team_id, user_to_delete.id)
        else:
            # current user is the global_admin.
            new_owner_id = user_to_delete.id

        model.user_id = new_owner_id

    session.bulk_save_objects(models)
