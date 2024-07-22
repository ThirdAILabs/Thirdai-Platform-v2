import json
import re
import subprocess
from typing import Dict, List

import docker
from azure.containerregistry import ContainerRegistryClient
from azure.identity import ClientSecretCredential
from cloud_provider_interface import CloudProviderInterface
from utils import image_name_for_branch


class AzureProvider(CloudProviderInterface):
    def __init__(self, registry: str):
        self.registry = registry
        self.registry_name = registry.split(".")[0]

    def login(self, username: str, password: str, registry: str) -> None:
        client = docker.from_env()
        client.login(username=username, password=password, registry=registry)

    def build_image(
        self, path: str, tag: str, nocache: bool, buildargs: Dict[str, str]
    ) -> str:
        print(f"Building image at path: {path} with tag: {tag}")
        docker_client = docker.APIClient(base_url="unix://var/run/docker.sock")
        generator = docker_client.build(
            path=path,
            tag=tag,
            rm=True,
            platform="linux/x86_64",
            nocache=nocache,
            buildargs=buildargs,
        )
        image_id = None
        for chunk in generator:
            for minichunk in chunk.strip(b"\r\n").split(b"\r\n"):
                json_chunk = json.loads(minichunk)
                if "stream" in json_chunk:
                    print(json_chunk["stream"].strip())
                    match = re.search(
                        r"(^Successfully built |sha256:)([0-9a-f]+)$",
                        json_chunk["stream"],
                    )
                    if match:
                        image_id = match.group(2)
                if "errorDetail" in json_chunk:
                    raise RuntimeError(json_chunk["errorDetail"]["message"])
        if not image_id:
            raise RuntimeError(f"Did not successfully build {tag} from {path}")

        print(f"\nLocal: Built {image_id}\n")

        print("\n===============================================================\n")

        return image_id

    def push_image(self, image_id: str, tag: str) -> None:
        client = docker.from_env()
        image = client.images.get(image_id)
        image.tag(tag)
        for line in client.images.push(tag, stream=True, decode=True):
            print(line)

    def get_image_digest(self, name: str, tag: str) -> List[str]:
        client = docker.from_env()
        image_full_name = f"{self.registry}/{name}:{tag}"
        try:
            image = client.images.pull(image_full_name)
            digest = image.attrs["RootFS"]["Layers"]
            return digest
        except docker.errors.ImageNotFound as e:
            print(f"{image_full_name} not found: {e}")
            return None

    def delete_image(self, repository: str, tag: str, **kwargs) -> None:
        credential = ClientSecretCredential(
            tenant_id=kwargs.get("tenant_id"),
            client_id=kwargs.get("client_id"),
            client_secret=kwargs.get("client_secret"),
        )
        registry_client = ContainerRegistryClient(self.registry, credential)
        registry_client.delete_tag(repository, tag)

    def create_credentials(
        self, name: str, image_names: List[str], push_access: bool
    ) -> Dict[str, str]:
        check_scope_map_cmd = (
            "az acr scope-map show"
            f" --name {name}"
            f" --registry {self.registry_name}"
            " --output json"
        )
        p = subprocess.Popen(
            [check_scope_map_cmd],
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            shell=True,
        )
        out = p.stdout.read()
        if out:
            raise Exception(
                f"Scope map with the given name {name} already exists. Please reuse those credentials instead, or use a new name."
            )

        check_token_cmd = (
            "az acr token show"
            f" --name {name}"
            f" --registry {self.registry_name}"
            " --output json"
        )
        p = subprocess.Popen(
            [check_token_cmd],
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            shell=True,
        )
        out = p.stdout.read()
        if out:
            raise Exception(
                f"Token with the given name {name} already exists. Please reuse those credentials instead, or use a new name."
            )

        make_scope_map_cmd = (
            "az acr scope-map create"
            f" --name {name}"
            f" --registry {self.registry_name}"
            " --output json"
        )
        for image_name in image_names:
            make_scope_map_cmd += (
                f" --repository {image_name} content/read metadata/read"
            )
            if push_access:
                make_scope_map_cmd += " content/write metadata/write content/delete"
        p = subprocess.Popen([make_scope_map_cmd], stdout=subprocess.PIPE, shell=True)
        out = p.stdout.read()
        print(out)

        make_token_cmd = (
            "az acr token create"
            f" --name {name}"
            f" --registry {self.registry_name}"
            f" --scope-map {name}"
            " --output json"
        )
        p = subprocess.Popen([make_token_cmd], stdout=subprocess.PIPE, shell=True)
        out = p.stdout.read()
        out = json.loads(out)
        username = out["credentials"]["username"]
        password = out["credentials"]["passwords"][0]["value"]

        return {"username": username, "password": password}

    def update_credentials(
        self, name: str, image_names: List[str], push_access: bool
    ) -> None:
        check_scope_map_cmd = (
            "az acr scope-map show"
            f" --name {name}"
            f" --registry {self.registry_name}"
            " --output json"
        )
        p = subprocess.Popen(
            [check_scope_map_cmd],
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            shell=True,
        )
        out = p.stdout.read()
        if not out:
            raise Exception(
                f"Scope map with the given name {name} does not exist. Please first create a scope map named {name}."
            )

        check_token_cmd = (
            "az acr token show"
            f" --name {name}"
            f" --registry {self.registry_name}"
            " --output json"
        )
        p = subprocess.Popen(
            [check_token_cmd],
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            shell=True,
        )
        out = p.stdout.read()
        if not out:
            raise Exception(
                f"Token with the given name {name} does not exist. Please first create a token named {name}."
            )

        update_scope_map_cmd = (
            "az acr scope-map update"
            f" --name {name}"
            f" --registry {self.registry_name}"
            " --output json"
        )
        for image_name in image_names:
            update_scope_map_cmd += (
                f" --add-repository {image_name} content/read metadata/read"
            )
            if push_access:
                update_scope_map_cmd += " content/write metadata/write content/delete"
        p = subprocess.Popen([update_scope_map_cmd], stdout=subprocess.PIPE, shell=True)
        out = p.stdout.read()
        print(out)

    def update_image(self, image_id: str, name: str, tag: str) -> None:
        client = docker.from_env()
        image = client.images.get(image_id)
        image.tag(name, tag)
        for line in client.images.push(name, stream=True, decode=True):
            print(line)

    def get_registry_name(self) -> str:
        return self.registry

    def get_full_image_name(self, base_name: str, branch: str, tag: str) -> str:
        image_name = f"{self.registry}/{image_name_for_branch(base_name, branch)}:{tag}"
        return image_name
