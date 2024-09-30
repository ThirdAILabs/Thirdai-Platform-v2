import os
from urllib.parse import urlencode, urljoin

from backend.mailer import Mailer


def send_verification_mail(self, email: str, verification_token: str, username: str):
    """
    Send verification email to the user.
    """
    subject = "Verify Your Email Address"
    base_url = os.getenv("PUBLIC_MODEL_BAZAAR_ENDPOINT")
    args = {"verification_token": verification_token}
    verify_link = urljoin(base_url, f"api/user/redirect-verify?{urlencode(args)}")
    body = f"<p>Please click the following link to verify your email address: <a href='{verify_link}'>verify</a></p>"

    Mailer(to=f"{username} <{email}>", subject=subject, body=body)
