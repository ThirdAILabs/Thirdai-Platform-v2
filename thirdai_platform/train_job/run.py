try:
    import logging
    import sys
    from pathlib import Path

    import nltk
    from licensing.verify import verify_license

    nltk.download("punkt_tab")
    print("Downloading punkttab")

    import argparse

    from platform_common.logging import LogCode, TrainingLogger
    from platform_common.pydantic_models.training import (
        ModelType,
        TrainConfig,
        UDTSubType,
    )
    from train_job.models.classification_models import (
        TextClassificationModel,
        TokenClassificationModel,
    )
    from train_job.models.neural_db_v2 import NeuralDBV2
    from train_job.reporter import HttpReporter, Reporter
except ImportError as e:
    logging.error(f"Failed to import module: {e}")
    sys.exit(f"ImportError: {e}")


def get_model(config: TrainConfig, reporter: Reporter, logger: TrainingLogger):
    model_type = config.model_options.model_type

    if model_type == ModelType.NDB:
        logger.info("Creating NDB model", code=LogCode.MODEL_INIT)
        return NeuralDBV2(config, reporter, logger)
    elif model_type == ModelType.UDT:
        udt_type = config.model_options.udt_options.udt_sub_type
        logger.info(f"UDT type: {udt_type}", code=LogCode.MODEL_INIT)
        if udt_type == UDTSubType.text:
            return TextClassificationModel(config, reporter, logger)
        elif udt_type == UDTSubType.token:
            return TokenClassificationModel(config, reporter, logger)
        else:
            message = f"Unsupported UDT subtype '{udt_type.value}'"
            logger.error(message, code=LogCode.MODEL_INIT)
            raise ValueError(message)

    message = f"Unsupported model type {model_type.value}"
    logger.error(message, code=LogCode.MODEL_INIT)
    raise ValueError(message)


def load_config():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str, required=True)

    args = parser.parse_args()

    with open(args.config) as file:
        return TrainConfig.model_validate_json(file.read())


def main():
    try:
        config: TrainConfig = load_config()
        log_dir: Path = Path(config.model_bazaar_dir) / "logs" / config.model_id

        logger = TrainingLogger(
            log_dir=log_dir,
            log_prefix="train",
            model_id=config.model_id,
            model_type=config.model_options.model_type,
            user_id=config.user_id,
        )

        reporter = HttpReporter(config.model_bazaar_endpoint, logger)

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
