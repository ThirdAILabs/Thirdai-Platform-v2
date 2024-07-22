from pydantic import BaseModel


def image_name_for_branch(name: str, branch: str) -> str:
    return f"{name}_{branch}" if branch != "prod" else name


class Credentials(BaseModel):
    push_username: str
    pull_username: str
    push_password: str
    pull_password: str
