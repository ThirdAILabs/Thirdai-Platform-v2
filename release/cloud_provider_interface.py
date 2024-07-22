from typing import Dict, List


class CloudProviderInterface:
    def login(self, username: str, password: str, registry: str) -> None:
        raise NotImplementedError

    def build_image(
        self, path: str, tag: str, nocache: bool, buildargs: Dict[str, str]
    ) -> str:
        raise NotImplementedError

    def push_image(self, image_id: str, tag: str) -> None:
        raise NotImplementedError

    def get_image_digest(self, name: str, tag: str) -> List[str]:
        raise NotImplementedError

    def delete_image(self, repository: str, tag: str, **kwargs) -> None:
        raise NotImplementedError

    def create_credentials(
        self, name: str, image_names: List[str], push_access: bool
    ) -> Dict[str, str]:
        raise NotImplementedError

    def update_credentials(
        self, name: str, image_names: List[str], push_access: bool
    ) -> None:
        raise NotImplementedError

    def update_image(self, image_id: str, name: str, tag: str) -> None:
        raise NotImplementedError

    def get_registry_name(self) -> str:
        raise NotImplementedError

    def get_full_image_name(self, base_name: str, branch: str, tag: str) -> str:
        raise NotImplementedError
