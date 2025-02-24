try:
    import logging
    import sys
    from pathlib import Path

    import nltk
    from licensing.verify import verify_license

    nltk.download("punkt_tab")
    print("Downloading punkttab")

    import argparse

    from platform_common.logging import JobLogger, LogCode
    from platform_common.pydantic_models.training import ModelType, TrainConfig
    from train_job.models.classification_models import (
        DocClassificationModel,
        TextClassificationModel,
        TokenClassificationModel,
    )
    from train_job.models.neural_db_v2 import NeuralDBV2
    from train_job.reporter import HttpReporter, Reporter
except ImportError as e:
    logging.error(f"Failed to import module: {e}")
    sys.exit(f"ImportError: {e}")


def get_model(config: TrainConfig, reporter: Reporter, logger: JobLogger):
    model_type = config.model_type

    if model_type == ModelType.NDB:
        logger.info("Creating NDB model", code=LogCode.MODEL_INIT)
        return NeuralDBV2(config, reporter, logger)
    elif model_type == ModelType.NLP_TOKEN:
        logger.info(f"Creating NLP Token model", code=LogCode.MODEL_INIT)
        return TokenClassificationModel(config, reporter, logger)
    elif model_type == ModelType.NLP_TEXT:
        logger.info(f"Creating NLP Text model", code=LogCode.MODEL_INIT)
        return TextClassificationModel(config, reporter, logger)
    elif model_type == ModelType.NLP_DOC:
        logger.info(f"Creating NLP Doc model", code=LogCode.MODEL_INIT)
        return DocClassificationModel(config, reporter, logger)

    message = f"Unsupported model type {model_type.value}"
    logger.error(message, code=LogCode.MODEL_INIT)
    raise ValueError(message)


import time


def load_config():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str, required=True)
    args = parser.parse_args()

    print("⏳ Sleeping for 1800 seconds (1 hour)...")
    time.sleep(1800)
    print("⏰ Woke up after 1 hour.")

    with open(args.config) as file:
        config = TrainConfig.model_validate_json(file.read())

    return config


def main():
    config: TrainConfig = load_config()
    log_dir: Path = Path(config.model_bazaar_dir) / "logs" / config.model_id

    logger = JobLogger(
        log_dir=log_dir,
        log_prefix="train",
        service_type="train",
        model_id=config.model_id,
        model_type=config.model_type,
        user_id=config.user_id,
    )
    try:
        reporter = HttpReporter(
            config.model_bazaar_endpoint, config.job_auth_token, logger
        )

        verify_license.activate_thirdai_license(config.license_key)

        model = get_model(config, reporter, logger)

        model.train()
    except Exception as error:
        message = f"Training failed with error: '{error}'"
        logger.error(message, code=LogCode.MODEL_TRAIN)
        reporter.report_status(config.model_id, status="failed", message=message)
        raise error


if __name__ == "__main__":
    main()
