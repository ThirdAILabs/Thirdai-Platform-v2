import argparse
from pathlib import Path
from typing import Dict

import yaml
from azure_provider import AzureProvider
from cloud_provider_interface import CloudProviderInterface
from docker_constants import images_to_build, images_to_pull_from_private
from utils import Credentials, image_name_for_branch, load_config


def get_args() -> argparse.Namespace:
    """
    Parse and return command-line arguments.

    :return: Parsed arguments
    """
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-b",
        "--branch",
        required=True,
        help="The branch to push docker images to. E.g. 'prod', 'test', etc.",
    )
    parser.add_argument(
        "--config",
        default="config.yaml",
        help="Path to the YAML configuration file. Defaults to 'config.yaml' in the current directory if not specified.",
    )
    parser.add_argument(
        "--no-cache",
        action="store_true",
        help="If this flag is present, Docker will not use cache when building images.",
    )
    parser.add_argument(
        "--version",
        type=str,
        help="If provided it will use this version for the provided branch.",
    )
    parser.add_argument(
        "--dont-update-latest",
        action="store_true",
        help="If this flag is present, the 'latest' tag will not be updated.",
    )
    parser.add_argument(
        "--dont-update-scope",
        action="store_true",
        help="If this flag is present, we dont update the scope with latest images, helpful for running docker tests.",
    )
    args = parser.parse_args()
    print(f"[DEBUG] Parsed arguments: {args}")
    return args


def get_tag(branch: str, config: dict) -> str:
    """
    Get the version tag for a specific branch from the configuration.

    :param branch: Branch name
    :param config: Configuration dictionary
    :return: Version tag
    """
    tag = "v" + config[config["provider"]]["branches"][branch]["version"]
    print(f"[DEBUG] get_tag: branch={branch}, tag={tag}")
    return tag


def get_root_absolute_path() -> Path:
    """
    Get the absolute path to the project root.

    :return: Absolute path to the project root
    """
    current_path = Path(__file__).resolve()
    print(f"[DEBUG] Starting search for project root from: {current_path}")
    while current_path.name != "Thirdai-Platform-v2":
        current_path = current_path.parent
    print(f"[DEBUG] Found project root: {current_path}")
    return current_path


def build_image(
    provider: CloudProviderInterface,
    name: str,
    branch: str,
    tag: str,
    buildargs: Dict[str, str],
    nocache: bool,
    dockerfile_path: str,
    context_path: str,
) -> Dict[str, str]:
    """
    Build a Docker image.

    :param provider: Cloud provider interface
    :param name: Name of the image
    :param branch: Branch name
    :param tag: Version tag
    :param buildargs: Build arguments for Docker
    :param nocache: Whether to use cache during build
    :param dockerfile_path: Path to the actual Dockerfile
    :param context_path: Path to the context used to build
    :return: Dictionary with image name and image ID
    """
    print(f"[DEBUG] Starting build_image for {name}")
    dockerfile_path = Path(dockerfile_path)
    context_path = Path(context_path)
    if not context_path.is_absolute():
        context_path = get_root_absolute_path() / context_path
        print(f"[DEBUG] Updated context_path to absolute: {context_path}")
    if not dockerfile_path.is_absolute():
        dockerfile_path = context_path / Path(dockerfile_path)
        print(f"[DEBUG] Updated dockerfile_path to absolute: {dockerfile_path}")

    full_name = provider.get_full_image_name(name, branch, tag)
    print(f"[DEBUG] Full image name: {full_name}")
    image_id = provider.build_image(
        str(dockerfile_path), str(context_path), full_name, nocache, buildargs
    )
    print(f"[DEBUG] Built image {name} with ID: {image_id}")
    return {name: image_id}


def build_images(
    provider: CloudProviderInterface,
    branch: str,
    tag: str,
    username: str,
    password: str,
    nocache: bool,
) -> Dict[str, str]:
    """
    Build all Docker images.

    :param provider: Cloud provider interface
    :param branch: Branch name
    :param tag: Version tag
    :param username: Docker username
    :param password: Docker password
    :param nocache: Whether to use cache during build
    :return: Dictionary of image names and their IDs
    """
    print(f"[DEBUG] Starting build_images for branch {branch} with tag {tag}")
    image_ids = {}

    for image in images_to_build:
        print(f"[DEBUG] Processing image: {image.name}")
        buildargs = {}
        if image.name == "thirdai_platform":
            buildargs = {
                "tag": tag,
                "docker_registry": provider.get_registry_name(),
                "docker_username": username,
                "docker_password": password,
                **{
                    image.key: image_name_for_branch(image.name, branch)
                    for image in images_to_build
                },
            }
            print(f"[DEBUG] Build args for thirdai_platform: {buildargs}")

        image_ids.update(
            build_image(
                provider,
                image.name,
                branch,
                tag,
                buildargs,
                nocache,
                image.dockerfile_path,
                image.context_path,
            )
        )
        print(f"[DEBUG] Updated image_ids: {image_ids}")

    return image_ids


def verify_tag(
    provider: CloudProviderInterface, image_ids: Dict[str, str], tag: str, branch: str
) -> None:
    """
    Verify that the Docker image tag matches the checksum of the built image.

    :param provider: Cloud provider interface
    :param image_ids: Dictionary of image names and their IDs
    :param tag: Version tag
    :param branch: Branch name
    :raises RuntimeError: If an image with the same tag but different checksum exists
    """
    print(f"[DEBUG] Starting verify_tag for tag {tag} on branch {branch}")
    for name, image_id in image_ids.items():
        print(f"[DEBUG] Verifying image: {name} with local image ID: {image_id}")
        existing = provider.get_image_digest(
            name=image_name_for_branch(name, branch), tag=tag
        )
        new = provider.get_local_image_digest(image_id=image_id)
        print(f"[DEBUG] Existing digest: {existing}, New digest: {new}")
        if existing and existing != new:
            raise RuntimeError(
                f"A docker image with name '{name}' and branch '{branch}' and tag '{tag}' with "
                "a different checksum exists in the registry."
            )
    print(f"[DEBUG] verify_tag completed successfully.")


def push_images(
    provider: CloudProviderInterface,
    image_ids: Dict[str, str],
    tag: str,
    branch: str,
    dont_update_latest: bool,
) -> None:
    """
    Push Docker images to the registry.

    :param provider: Cloud provider interface
    :param image_ids: Dictionary of image names and their IDs
    :param tag: Version tag
    :param branch: Branch name
    :param dont_update_latest: Whether to update the 'latest' tag
    """
    print(f"[DEBUG] Starting push_images for branch {branch} with tag {tag}")
    for name, image_id in image_ids.items():
        full_image_name_tagged = provider.get_full_image_name(name, branch, tag)
        print(f"[DEBUG] Pushing image {name} with tag {tag}: {full_image_name_tagged}")
        provider.push_image(image_id, full_image_name_tagged)
        if not dont_update_latest:
            full_image_name_latest = provider.get_full_image_name(
                name, branch, "latest"
            )
            print(
                f"[DEBUG] Also pushing image {name} with tag latest: {full_image_name_latest}"
            )
            provider.push_image(image_id, full_image_name_latest)
    print(f"[DEBUG] push_images completed.")


def save_config(config_path: str, config: dict) -> None:
    """
    Save the configuration dictionary to a YAML file.

    :param config_path: Path to the configuration file
    :param config: Configuration dictionary
    """
    print(f"[DEBUG] Saving configuration to {config_path}")
    with open(config_path, "w") as file:
        yaml.dump(config, file, default_flow_style=False)
    print(f"[DEBUG] Configuration saved successfully.")


def main() -> None:
    """
    Main function to build and push Docker images.
    """
    print("[DEBUG] Starting main execution.")
    args = get_args()
    config = load_config(args.config)
    print(f"[DEBUG] Loaded config: {config}")

    provider_name = config["provider"]
    if provider_name == "azure":
        if not config["azure"]["registry"]:
            config["azure"]["registry"] = "thirdaiplatform.azurecr.io"
            print("[DEBUG] Set default registry for azure.")

        # Ensure branch configuration exists
        if "branches" not in config["azure"]:
            config["azure"]["branches"] = {}
            print("[DEBUG] Added missing 'branches' to azure config.")

        if args.branch not in config["azure"]["branches"]:
            config["azure"]["branches"][args.branch] = {
                "version": "0.0.1",
                "push_credentials": {"username": "", "password": ""},
                "pull_credentials": {"username": "", "password": ""},
            }
            print(f"[DEBUG] Added branch configuration for {args.branch}.")

        tag = "v" + args.version if args.version else get_tag(args.branch, config)
        print(f"[DEBUG] Using tag: {tag}")

        azure_config = config["azure"]
        branch_config = azure_config["branches"][args.branch]
        provider = AzureProvider(registry=azure_config["registry"])
        print(
            f"[DEBUG] Initialized AzureProvider with registry: {azure_config['registry']}"
        )

        push_credentials = branch_config.get("push_credentials", {})
        pull_credentials = branch_config.get("pull_credentials", {})
        push_username = push_credentials.get("username")
        push_password = push_credentials.get("password")
        pull_username = pull_credentials.get("username")
        pull_password = pull_credentials.get("password")

        # Replace underscores with hyphens in the branch name
        sanitized_branch = args.branch.replace("_", "-")
        print(f"[DEBUG] Sanitized branch name: {sanitized_branch}")

        if not push_username or not push_password:
            print("[DEBUG] Creating new push credentials.")
            new_push_credentials = provider.create_credentials(
                name=f"thirdaiplatform-push-{sanitized_branch}",
                image_names=[
                    image_name_for_branch(image.name, args.branch)
                    for image in images_to_build
                ]
                + images_to_pull_from_private,
                push_access=True,
            )
            push_username = new_push_credentials["username"]
            push_password = new_push_credentials["password"]

            config["azure"]["branches"][args.branch]["push_credentials"] = {
                "username": push_username,
                "password": push_password,
            }
            print(f"[DEBUG] New push credentials: {new_push_credentials}")
        else:
            if not args.dont_update_scope:
                print("[DEBUG] Updating push credentials scope.")
                provider.update_credentials(
                    name=f"thirdaiplatform-push-{sanitized_branch}",
                    image_names=[
                        image_name_for_branch(image.name, args.branch)
                        for image in images_to_build
                    ]
                    + images_to_pull_from_private,
                    push_access=True,
                )

        if not pull_username or not pull_password:
            print("[DEBUG] Creating new pull credentials.")
            new_pull_credentials = provider.create_credentials(
                name=f"thirdaiplatform-pull-{sanitized_branch}",
                image_names=[
                    image_name_for_branch(image.name, args.branch)
                    for image in images_to_build
                ]
                + images_to_pull_from_private,
                push_access=False,
            )
            pull_username = new_pull_credentials["username"]
            pull_password = new_pull_credentials["password"]

            config["azure"]["branches"][args.branch]["pull_credentials"] = {
                "username": pull_username,
                "password": pull_password,
            }
            print(f"[DEBUG] New pull credentials: {new_pull_credentials}")
        else:
            if not args.dont_update_scope:
                print("[DEBUG] Updating pull credentials scope.")
                provider.update_credentials(
                    name=f"thirdaiplatform-pull-{sanitized_branch}",
                    image_names=[
                        image_name_for_branch(image.name, args.branch)
                        for image in images_to_build
                    ]
                    + images_to_pull_from_private,
                    push_access=False,
                )

        # Write back the configuration to ensure it is up-to-date
        save_config(args.config, config)

        provider.authorize_credentials(
            credentials=Credentials(
                push_username=push_username,
                push_password=push_password,
                pull_username=pull_username,
                pull_password=pull_password,
            )
        )
        print("[DEBUG] Credentials authorized.")

        # Build images
        image_ids = build_images(
            provider, args.branch, tag, pull_username, pull_password, args.no_cache
        )
        print(f"[DEBUG] Built images: {image_ids}")

        verify_tag(provider, image_ids, tag, args.branch)
        print("[DEBUG] Tag verification successful.")

        push_images(provider, image_ids, tag, args.branch, args.dont_update_latest)
        print("[DEBUG] push_images completed.")
    print("[DEBUG] main execution completed.")


if __name__ == "__main__":
    main()
