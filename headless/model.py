import os
from typing import Dict, List, Tuple

from thirdai.neural_db import ModelBazaar

from headless.utils import get_csv_source_id


class Flow:
    def __init__(self, base_url: str, email: str, password: str):
        self._bazaar_client = ModelBazaar(base_url=base_url)
        self._bazaar_client.log_in(email=email, password=password)

    @property
    def bazaar_client(self):
        return self._bazaar_client

    def train(
        self,
        model_name: str,
        unsupervised_docs: List[str] = None,
        supervised_docs: List[Tuple[str, str]] = None,
        test_doc: str = None,
        doc_type: str = "local",
        extra_options: Dict = {},
        base_model_identifier: str = None,
        is_async: bool = True,
        metadata: List[Dict[str, str]] = None,
        nfs_base_path: str = None,
    ):
        print("*" * 50 + f" Training the model: {model_name} " + "*" * 50)
        if supervised_docs:
            if metadata is None:
                metadata = [None] * len(supervised_docs)
            supervised_tuple = [
                (
                    sup_file,
                    get_csv_source_id(
                        (
                            unsup_file
                            if not doc_type == "nfs"
                            else os.path.join(nfs_base_path, unsup_file[1:])
                        ),
                        extra_options.get("csv_id_column"),
                        extra_options.get("csv_strong_columns"),
                        extra_options.get("csv_weak_columns"),
                        extra_options.get("csv_reference_columns"),
                        file_metadata,
                    ),
                )
                for (sup_file, unsup_file), file_metadata in zip(
                    supervised_docs, metadata
                )
            ]
        else:
            supervised_tuple = []
        return self._bazaar_client.train(
            model_name=model_name,
            unsupervised_docs=unsupervised_docs,
            supervised_docs=supervised_tuple,
            test_doc=test_doc,
            doc_type=doc_type,
            train_extra_options=extra_options,
            base_model_identifier=base_model_identifier,
            is_async=is_async,
            metadata=metadata,
        )

    def deploy(
        self,
        model_identifier,
        deployment_name,
        is_async: bool = True,
    ):
        print("*" * 50 + f" Deploying the model {model_identifier} " + "*" * 50)
        return self._bazaar_client.deploy(
            model_identifier=model_identifier,
            deployment_name=deployment_name,
            is_async=is_async,
        )
