import argparse
import logging
from pathlib import Path

from platform_common.logging import setup_logger
from platform_common.pii.udt_common_patterns import find_common_pattern
from platform_common.pydantic_models.training import (
    DatagenConfig,
    ModelType,
    NlpTextDatagenOptions,
    NlpTokenDatagenOptions,
)

logger = logging.getLogger("data_generation")


def load_config():
    parser = argparse.ArgumentParser()
    parser.add_argument("--config", type=str, required=True)

    args = parser.parse_args()

    with open(args.config) as file:
        return DatagenConfig.model_validate_json(file.read())


def main():
    """
    Main function to initialize and generate the data based on environment variables.
    """

    config: DatagenConfig = load_config()
    print(f"{config = }", flush=True)

    log_dir: Path = Path(config.model_bazaar_dir) / "logs" / config.model_id
    setup_logger(log_dir=log_dir, log_prefix="data_generation")
    logger = logging.getLogger("data_generation")

    if config.task_options.model_type == ModelType.NLP_TEXT:
        from data_generation_job.text_data_factory import TextDataFactory

        factory = TextDataFactory(logger=logger, config=config)

        task_opts: NlpTextDatagenOptions = config.task_options

        dataset_config = factory.generate_data(
            task_prompt=config.task_prompt,
            samples_per_label=task_opts.samples_per_label,
            target_labels=task_opts.labels,
            user_vocab=task_opts.user_vocab,
            user_prompts=task_opts.user_prompts,
            vocab_per_sentence=task_opts.vocab_per_sentence,
        )

        logger.info(f"Text data generation initialized with args: {task_opts}")

        udt_options = {
            "udt_sub_type": "text",
            "text_column": dataset_config["input_feature"],
            "label_column": dataset_config["target_feature"],
            "n_target_classes": len(dataset_config["target_labels"]),
        }

    else:
        from data_generation_job.token_data_factory import TokenDataFactory

        factory = TokenDataFactory(logger=logger, config=config)

        task_opts: NlpTokenDatagenOptions = config.task_options

        common_patterns = [
            tag.name for tag in task_opts.tags if find_common_pattern(tag.name)
        ]

        dataset_config = factory.generate_data(
            task_prompt=config.task_prompt,
            tags=(
                [tag for tag in task_opts.tags if find_common_pattern(tag.name) is None]
            ),
            num_sentences_to_generate=task_opts.num_sentences_to_generate,
            num_samples_per_tag=task_opts.num_samples_per_tag,
            samples=task_opts.samples,
            templates_per_sample=task_opts.templates_per_sample,
            load_from_storage=task_opts.load_from_storage,
        )

        logger.info(
            f"Token data generation initialized with args: {task_opts} and common_patterns: {common_patterns}"
        )
        print(f"{dataset_config = }", flush=True)
        udt_options = {
            "udt_sub_type": "token",
            "source_column": dataset_config["input_feature"],
            "target_column": dataset_config["target_feature"],
            "target_labels": dataset_config["target_labels"] + common_patterns,
        }

    logger.info(f"Prepared UDT options for training job: {udt_options}")


if __name__ == "__main__":
    main()
