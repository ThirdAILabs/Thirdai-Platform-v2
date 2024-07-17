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


def await_train(inputs: Dict):
    logging.info(f"inputs: {inputs}")
    model = inputs.get("model")
    logging.info(
        f"Waiting for training to finish for model {model.model_identifier} and id {model.model_id}"
    )
    flow.bazaar_client.await_train(model)


def deploy(inputs: Dict):
    logging.info(f"inputs: {inputs}")
    model = inputs.get("model")
    run_name = inputs.get("run_name")

    logging.info(
        f"Deploying the model {model.model_identifier} and id {model.model_id}"
    )

    return flow.deploy(model.model_identifier, f"{run_name}_deployment")


def await_deploy(inputs: Dict):
    logging.info(f"inputs: {inputs}")
    deployment = inputs.get("deployment")
    logging.info(
        f"Waiting for Deployment to finish for deployment {deployment.deployment_identifier}"
    )
    flow.bazaar_client.await_deploy(deployment)


def check_deployment(inputs: Dict):
    logging.info(f"inputs: {inputs}")
    deployment = inputs.get("deployment")
    run_name = inputs.get("run_name")

    logging.info(f"checking the deployment for {deployment.deployment_identifier}")

    logging.info("Searching the deployment")
    results = deployment.search(
        query="Can autism and down syndrome be in conjunction",
        top_k=5,
    )

    query_text = results["query_text"]
    references = results["references"]

    best_answer = references[4]
    good_answer = references[2]

    logging.info(f"upvoting the model")
    deployment.upvote(
        [
            {"query_text": query_text, "reference_id": best_answer["id"]},
            {"query_text": query_text, "reference_id": good_answer["id"]},
        ]
    )

    logging.info("Associating the model")
    deployment.associate(
        [
            {"source": "authors", "target": "contributors"},
            {"source": "paper", "target": "document"},
        ]
    )

    logging.info("Checking the sources")
    deployment.sources()

    logging.info("Ovveriding the model")
    deployment.save_model(override=True)


def undeploy(inputs: Dict):
    logging.info(f"inputs: {inputs}")
    deployment = inputs.get("deployment")

    logging.info(f"stopping the deployment for {deployment.deployment_identifier}")

    flow.bazaar_client.undeploy(deployment)


functions_registry = {
    "check_unsupervised": check_unsupervised,
    "check_unsupervised_supervised": check_unsupervised_supervised,
    "await_train": await_train,
    "deploy": deploy,
    "check_deployment": check_deployment,
    "await_deploy": await_deploy,
    "undeploy": undeploy,
}
