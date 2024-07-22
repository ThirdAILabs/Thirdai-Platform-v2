def image_name_for_branch(name: str, branch: str) -> str:
    return f"{name}_{branch}" if branch != "prod" else name
