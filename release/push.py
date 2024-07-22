import argparse
import os
from pathlib import Path
from typing import Dict

import yaml
from azure_provider import AzureProvider
from cloud_provider_interface import CloudProviderInterface
from docker_constants import image_base_names
from utils import Credentials, image_name_for_branch


def get_args():
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
    return parser.parse_args()


def get_tag(branch: str, config: dict):
    return "v" + config[config["provider"]]["branches"][branch]["version"]


def get_root_absolute_path():
    current_path = Path(__file__).resolve()
    while current_path.name != "ThirdAI-Platform":
        current_path = current_path.parent

    print(current_path)
    return current_path


def build_image(
    provider: CloudProviderInterface,
    name: str,
    branch: str,
    tag: str,
    buildargs: Dict[str, str],
    nocache: bool,
):
    path = get_root_absolute_path() / name
    full_name = provider.get_full_image_name(name, branch, tag)
    image_id = provider.build_image(str(path), full_name, nocache, buildargs)
    return {name: image_id}


def build_images(
    provider: CloudProviderInterface,
    branch: str,
    tag: str,
    username: str,
    password: str,
    nocache: bool,
):
    image_ids = {}

    # Build ThirdAI platform image with specific buildargs
    buildargs = {
        "tag": tag,
        "docker_registry": provider.get_registry_name(),
        "docker_username": username,
        "docker_password": password,
        "export_image_names_command": (
            " ".join(
                [
                    f"export {key}={image_name_for_branch(base_name, branch)}"
                    for key, base_name in image_base_names.peripherals_as_dict().items()
                ]
            )
        ),
    }
    image_ids.update(
        build_image(
            provider,
            image_base_names.THIRDAI_PLATFORM_IMAGE_NAME,
            branch,
            tag,
            buildargs,
            nocache,
        )
    )

    # Build peripheral images without buildargs
    for base_name in image_base_names.peripherals_as_dict().values():
        image_ids.update(build_image(provider, base_name, branch, tag, {}, nocache))

    return image_ids


def verify_tag(
    provider: CloudProviderInterface, image_ids: Dict[str, str], tag: str, branch: str
):
    for name, image_id in image_ids.items():
        existing = provider.get_image_digest(
            name=image_name_for_branch(name, branch), tag=tag
        )
        new = provider.get_local_image_digest(image_id=image_id)
        if existing and existing != new:
            raise RuntimeError(
                f"A docker image with name '{name}' and tag '{tag}' with "
                "a different checksum exists in the registry."
            )


def push_images(
    provider: CloudProviderInterface,
    image_ids: Dict[str, str],
    tag: str,
    branch: str,
    dont_update_latest: bool,
):
    for name, image_id in image_ids.items():
        provider.push_image(image_id, provider.get_full_image_name(name, branch, tag))
        if not dont_update_latest:
            provider.push_image(
                image_id, provider.get_full_image_name(name, branch, "latest")
            )


def load_config(config_path: str):
    if os.path.exists(config_path):
        with open(config_path, "r") as file:
            return yaml.safe_load(file)
    else:
        return {"provider": "azure", "azure": {"registry": "", "branches": {}}}


def save_config(config_path: str, config: dict):
    with open(config_path, "w") as file:
        yaml.dump(config, file, default_flow_style=False)


def main():
    args = get_args()
    config = load_config(args.config)

    provider_name = config["provider"]
    if provider_name == "azure":
        if not config["azure"]["registry"]:
            config["azure"]["registry"] = "thirdaiplatform.azurecr.io"

        # Ensure branch configuration exists
        if "branches" not in config["azure"]:
            config["azure"]["branches"] = {}

        if args.branch not in config["azure"]["branches"]:
            config["azure"]["branches"][args.branch] = {
                "version": "0.0.1",
                "push_credentials": {"username": "", "password": ""},
                "pull_credentials": {"username": "", "password": ""},
            }

        tag = "v" + args.version if args.version else get_tag(args.branch, config)

        azure_config = config["azure"]
        branch_config = azure_config["branches"][args.branch]
        provider = AzureProvider(registry=azure_config["registry"])

        push_credentials = branch_config.get("push_credentials", {})
        pull_credentials = branch_config.get("pull_credentials", {})
        push_username = push_credentials.get("username")
        push_password = push_credentials.get("password")
        pull_username = pull_credentials.get("username")
        pull_password = pull_credentials.get("password")

        # Replace underscores with hyphens in the branch name
        sanitized_branch = args.branch.replace("_", "-")

        if not push_username or not push_password:
            new_push_credentials = provider.create_credentials(
                name=f"thirdaiplatform-push-{sanitized_branch}",
                image_names=[
                    image_name_for_branch(name, args.branch)
                    for name in image_base_names.to_list()
                ],
                push_access=True,
            )
            push_username = new_push_credentials["username"]
            push_password = new_push_credentials["password"]

            config["azure"]["branches"][args.branch]["push_credentials"] = {
                "username": push_username,
                "password": push_password,
            }

        if not pull_username or not pull_password:
            new_pull_credentials = provider.create_credentials(
                name=f"thirdaiplatform-pull-{sanitized_branch}",
                image_names=[
                    image_name_for_branch(name, args.branch)
                    for name in image_base_names.to_list()
                ],
                push_access=False,
            )
            pull_username = new_pull_credentials["username"]
            pull_password = new_pull_credentials["password"]

            config["azure"]["branches"][args.branch]["pull_credentials"] = {
                "username": pull_username,
                "password": pull_password,
            }

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

        # Build images
        image_ids = build_images(
            provider, args.branch, tag, pull_username, pull_password, args.no_cache
        )

        verify_tag(provider, image_ids, tag, args.branch)
        push_images(provider, image_ids, tag, args.branch, args.dont_update_latest)


if __name__ == "__main__":
    main()
