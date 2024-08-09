import csv
import json
import os
import random
import traceback
from collections import defaultdict
from concurrent.futures import ProcessPoolExecutor, as_completed
from resource.token_prompts import (
    attribute_dimension_prompt,
    attribute_value_prompt,
    dataset_generation_prompt,
    tag_value_prompt,
    template_prompt,
)
from typing import Dict, List

from data_factory_interface import DataFactory
from faker import Faker
from tqdm import tqdm
from utils import (
    assert_sufficient_examples,
    fill_and_transform_templates,
    subsample_dictionary,
)


class TokenDataFactory(DataFactory):
    SOURCE_COLUMN = "source"
    TARGET_COLUMN = "target"

    def __init__(
        self,
    ):
        super().__init__()
        self.faker = Faker()

        # All methods present in the faker to generate forged tags. E.g: credit_card_expire(), credit_card_expire(), first_name(), language_name(), ..
        self.faked_methods = [
            method
            for provider in self.faker.providers
            for method in dir(provider)
            if not method.startswith("_")
        ]

    def get_attributes(self, domain_prompt: str):
        response: str = self.llm_model.completion(
            prompt=attribute_dimension_prompt.format(domain_prompt=domain_prompt),
        )
        attributes = response.split("\n")

        attribute_values = {}
        for attribute in tqdm(attributes, desc="Attributed definition...", leave=True):
            response = self.llm_model.completion(
                prompt=attribute_value_prompt.format(
                    domain_prompt=domain_prompt, attribute=attribute
                ),
            )
            attribute_values[attribute] = response.split("\n")

        return attribute_values

    def get_fake_tag_values(self, tag: str, num_samples: int):
        # NOTE: It could be better to have an exact match
        matched_attrs = list(
            filter(lambda method: tag.lower() in method.lower(), self.faked_methods)
        )
        if not matched_attrs:
            return []

        matched_attr = min(matched_attrs, key=len)

        return [self.faker.__getattr__(matched_attr)() for _ in range(num_samples)]

    def get_complete_tag_examples(
        self,
        tag_examples: Dict[str, List[str]],
        total_expected_sentences: int,
        num_samples_per_tag: int,
    ):
        complete_tag_examples = defaultdict(list)

        for tag, user_examples in tqdm(
            tag_examples.items(), desc="Generating Sample for attributes: ", leave=True
        ):
            # Adding user examples
            complete_tag_examples[tag].extend(user_examples)

            # Trying to generate more examples from faker
            samples = self.get_fake_tag_values(
                tag,
                num_samples=total_expected_sentences,
            )
            if samples:
                complete_tag_examples[tag].extend(samples)
                continue

            # Not able to generate by faker so, generating samples by llm
            sampled_user_examples = random.sample(
                user_examples, min(3, len(user_examples))
            )

            response = self.llm_model.completion(
                prompt=tag_value_prompt.format(
                    num_samples_per_tag=num_samples_per_tag,
                    tag=tag,
                    tag_example=str(sampled_user_examples),
                )
            )
            complete_tag_examples[tag].extend(response.split("\n"))

        return complete_tag_examples

    def get_templatized_examples(self, tags: List[str], k: int = 2):
        return self.llm_model.completion(
            template_prompt.format(tag=", ".join(tags).replace("'", ""), k=k)
        )

    def get_value_requirements(self, attribute_values: Dict[str, List[str]]):
        subsampled_dict = subsample_dictionary(attribute_values)
        sampled_keys = random.sample(
            [*subsampled_dict], k=min(10, len(subsampled_dict))
        )

        values_requirements = "Take inspiration from the ideas below but do not mimic them directly. Ensure your output revolves around similar topics with some variations for accuracy.\n"
        for key in sampled_keys:
            values = subsampled_dict[key]
            values_requirements += (
                f"Include the following {key}: {' and '.join(values)}.\n"
            )

        return values_requirements

    def generate_data(
        self,
        domain_prompt: str,
        tags: List[str],
        tag_examples: Dict[str, List[str]],
        num_call_batches: int,
        batch_size=40,
        num_samples_per_tag=4,
        sentences_generated=0,  # To resume the generate function incase of midway failure. TODO(Gautam): Incorporate resuming the data_generation task
    ):
        total_expected_sentences = batch_size * num_call_batches
        assert total_expected_sentences < sentences_generated, "Invalid configuration"

        assert_sufficient_examples(tags, tag_examples)

        def process_task(prompt: str):
            response = self.llm_model.completion(
                prompt,
                system_prompt=f"You are a helpful assistant designed to generate synthetic data for domain {domain_prompt}.",
            )
            return response.split("\n")

        attribute_values = self.get_attributes(domain_prompt)

        complete_tag_examples = self.get_complete_tag_examples(
            tag_examples, total_expected_sentences, num_samples_per_tag
        )

        templatized_sentences_examples = self.get_templatized_examples(
            tags=random.choices(tags, k=min(10, tags))
        )

        tasks = []
        for _ in range(num_call_batches):
            # TODO(anyone): we should also add the [user_tag -> examples] in dataset_generation_prompt.
            random_prompts = self.get_random_prompts()

            values_requirements = self.get_value_requirements(attribute_values)
            tasks.append(
                dataset_generation_prompt.format(
                    domain_prompt=domain_prompt,
                    batch_size=batch_size,
                    sampled_tags=random.sample(tags, k=min(5, len(tags))),
                    templatized_sentences_examples=templatized_sentences_examples,
                    rnd_prompts_str="\n -\t".join(random_prompts),
                    values_requirements=values_requirements,
                )
            )

        random.shuffle(tasks)
        tasks = tasks[: total_expected_sentences - sentences_generated]

        write_chunk_size = 50
        total_chunks = len(tasks) // write_chunk_size + 1
        for idx in range(0, len(tasks), write_chunk_size):
            chunk = tasks[idx : idx + write_chunk_size]
            generated_templates = []
            with ProcessPoolExecutor() as executor, tqdm(
                total=len(chunk),
                desc=f"Generating token data {(idx // write_chunk_size)}/{total_chunks}",
            ) as pbar:
                futures = []

                # Submit input_data to the executor
                for task in chunk:
                    future = executor.submit(process_task, task)
                    future.add_done_callback(lambda p: pbar.update())
                    futures.append(future)

                # Wait for all input_data to complete and handle exceptions
                for future in as_completed(futures):
                    try:
                        response = future.result()
                        if response:
                            generated_templates.append(response)

                    except Exception as e:
                        with open(self.errored_file_location, mode="a") as errored_fp:
                            traceback.print_exc(file=errored_fp)
                            errored_fp.write("\n" + "=" * 100 + "\n")

            data_points = fill_and_transform_templates(
                tags, generated_templates, complete_tag_examples
            )

            self.write_on_training_file(
                data_points,
                fieldnames=[
                    TokenDataFactory.SOURCE_COLUMN,
                    TokenDataFactory.TARGET_COLUMN,
                ],
                write_fields=sentences_generated == 0,
            )

            sentences_generated += len(data_points)

        dataset_config = {
            "filepath": self.train_file_location,
            "error_file": self.errored_file_location,
            "task": "TOKEN_CLASSIFICATION",
            "input_feature": TokenDataFactory.SOURCE_COLUMN,
            "target_feature": TokenDataFactory.TARGET_COLUMN,
            "target_labels": tags,
            "num_samples": sentences_generated,
        }
        self.save_dict(self.config_file_location, dataset_config)

        return dataset_config
