try:
    import logging
    import sys
    from pathlib import Path

    import nltk
    from licensing.verify import verify_license

    print("Downloading punkt_tab")
    nltk.download("punkt_tab")

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
    print(f"Model type: {model_type}")

    if model_type == ModelType.NDB:
        logger.info("Creating NDB model", code=LogCode.MODEL_INIT)
        print("Creating NDB model")
        return NeuralDBV2(config, reporter, logger)
    elif model_type == ModelType.NLP_TOKEN:
        logger.info(f"Creating NLP Token model", code=LogCode.MODEL_INIT)
        print("Creating NLP Token model")
        return TokenClassificationModel(config, reporter, logger)
    elif model_type == ModelType.NLP_TEXT:
        logger.info(f"Creating NLP Text model", code=LogCode.MODEL_INIT)
        print("Creating NLP Text model")
        return TextClassificationModel(config, reporter, logger)
    elif model_type == ModelType.NLP_DOC:
        logger.info(f"Creating NLP Doc model", code=LogCode.MODEL_INIT)
        print("Creating NLP Doc model")
        return DocClassificationModel(config, reporter, logger)

    message = f"Unsupported model type {model_type.value}"
    logger.error(message, code=LogCode.MODEL_INIT)
    print(message)
    raise ValueError(message)


def load_config():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str, required=True)
    print("Parsing arguments")

    args = parser.parse_args()
    print(f"Config file path: {args.config}")

    with open(args.config) as file:
        print("Loading configuration")
        return TrainConfig.model_validate_json(file.read())


def main():
    print("Starting main function")
    config: TrainConfig = load_config()
    log_dir: Path = Path(config.model_bazaar_dir) / "logs" / config.model_id
    print(f"Log directory: {log_dir}")

    logger = JobLogger(
        log_dir=log_dir,
        log_prefix="train",
        service_type="train",
        model_id=config.model_id,
        model_type=config.model_type,
        user_id=config.user_id,
    )
    try:
        print("Initializing reporter")
        reporter = HttpReporter(
            config.model_bazaar_endpoint, config.job_auth_token, logger
        )

        print("Activating license")
        verify_license.activate_thirdai_license(config.license_key)

        print("Getting model")
        model = get_model(config, reporter, logger)

        print("Starting training")
        model.train()
    except Exception as error:
        message = f"Training failed with error: '{error}'"
        logger.error(message, code=LogCode.MODEL_TRAIN)
        print(message)
        reporter.report_status(config.model_id, status="failed", message=message)
        raise error


if __name__ == "__main__":
    main()
