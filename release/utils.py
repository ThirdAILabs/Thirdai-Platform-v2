from pydantic import BaseModel


def image_name_for_branch(name: str, branch: str) -> str:
    """
    Generate the image name for a given branch.

    :param name: Base name of the image
    :param branch: Branch name
    :return: Image name with branch suffix, or base name if branch is 'prod'
    """
    return f"{name}_{branch}" if branch != "prod" else name


class Credentials(BaseModel):
    """
    Model to store credentials for Docker registry access.
    """

    push_username: str
    pull_username: str
    push_password: str
    pull_password: str
