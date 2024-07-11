import logging
import os
from typing import Dict

from headless.configs import Config
from headless.model import Flow
from headless.utils import build_extra_options

logging.basicConfig(level=logging.INFO)

flow = None


def initialize_flow(base_url, email, password):
    global flow
    flow = Flow(base_url=base_url, email=email, password=password)


def check_unsupervised(inputs: Dict):
    logging.info(f"Running unsupervised with {inputs}")
    sharded = inputs.get("sharded")
    run_name = inputs.get("run_name")
    config: Config = inputs.get("config")

    type = "single" if not sharded else "multiple"
    return flow.train(
        model_name=f"{run_name}_{config.name}_{type}_unsupervised",
        unsupervised_docs=[
            os.path.join(config.base_path, config.unsupervised_paths[0])
        ],
        extra_options=build_extra_options(config, sharded),
        doc_type=config.doc_type,
        nfs_base_path=config.nfs_original_base_path,
    )


def check_unsupervised_supervised(inputs: Dict):
    logging.info(f"Running unsupervised supervised with {inputs}")
    sharded = inputs.get("sharded")
    run_name = inputs.get("run_name")
    config: Config = inputs.get("config")

    type = "single" if not sharded else "multiple"
    return flow.train(
        model_name=f"{run_name}_{config.name}_{type}_unsupervised_supervised",
        unsupervised_docs=[
            os.path.join(config.base_path, config.unsupervised_paths[0])
        ],
        supervised_docs=[
            (
                os.path.join(config.base_path, config.supervised_paths[0]),
                os.path.join(config.base_path, config.unsupervised_paths[0]),
            )
        ],
        extra_options=build_extra_options(config, sharded),
        doc_type=config.doc_type,
        nfs_base_path=config.nfs_original_base_path,
    )


functions_registry = {
    "check_unsupervised": check_unsupervised,
    "check_unsupervised_supervised": check_unsupervised_supervised,
}
