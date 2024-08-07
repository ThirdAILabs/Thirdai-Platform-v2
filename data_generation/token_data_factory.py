import json
import os
import random
import traceback
from collections import defaultdict
from resource.token_prompts import (
    attribute_dimension_prompt,
    attribute_value_prompt,
    tag_value_prompt,
    template_prompt
)
from typing import Dict, List

from data_factory_interface import DataFactory
from faker import Faker
from tqdm import tqdm
from utils import fill_and_transform_templates, load_and_write_csv, subsample_dictionary


def assert_sufficient_examples(tags: List[str], examples: Dict[str, List[str]]):
    missing_examples = [label for label in tags if label not in examples]
    if missing_examples:
        raise ValueError(
            f"Examples are not given for all tags. Tags with missing examples: {', '.join(missing_examples)}"
        )


class TokenDataFactory(DataFactory):
    def __init__(
        self,
        api_key: str,
    ):
        super().__init__(api_key)

        self.faker = Faker()

        # All methods present in the faker to generate forged tags. E.g: credit_card_expire(), credit_card_expire(), first_name(), language_name(), ..
        self.faked_methods = [
            method
            for provider in self.faker.providers
            for method in dir(provider)
            if not method.startswith("_")
        ]

    def get_fake_tags(self, tag: str, num_samples: int):
        # NOTE: It could be better to have an exact match
        matched_attrs = list(
            filter(lambda method: tag.lower() in method.lower(), self.faked_methods)
        )
        if not matched_attrs:
            return []

        matched_attr = min(matched_attrs, key=len)

        return [self.faker.__getattr__(matched_attr)() for _ in range(num_samples)]

    def generate_data(
        self,
        domain_prompt: str,
        tags: List[str],
        num_call_batches: int,
        tag_examples: Dict[str, List[str]],
        batch_size=40,
        num_samples_per_tag=100,
        sentences_generated=0,  # To resume the generate function incase of midway failure. TODO(Gautam): Incorporate resuming the data_generation task
    ):
        total_expected_sentences = batch_size * num_call_batches
        assert total_expected_sentences < sentences_generated, "Invalid configuration"

        def process_task(prompt):
            try:
                response = self.llm_completion(
                    prompt,
                    system_prompt=f"You are a helpful assistant designed to generate synthetic data for domain {domain_prompt}.",
                )
                return response, None
            except Exception as e:
                print(f"An error occurred while generation: {str(e)}")
                print(traceback.format_exc())
                return None, error_msg

        assert_sufficient_examples(tags, tag_examples)

        response: str = self.llm_completion(
            prompt=attribute_dimension_prompt.format(domain_prompt=domain_prompt),
        )
        attributes = response.split("\n")

        attribute_values = {}
        for attribute in tqdm(attributes, desc="Attributed definition...", leave=True):
            response = self.llm_completion(
                prompt=attribute_value_prompt.format(
                    domain_prompt=domain_prompt, attribute=attribute
                ),
            )
            attribute_values[attribute] = response.split("\n")

        generated_samples = defaultdict(list)

        ## Generate sample for entities
        for tag, examples in tqdm(
            tag_examples.items(), desc="Generating Sample for attributes: "
        ):
            samples = self.get_fake_tags(
                tag,
                num_samples=total_expected_sentences,
            )
            if samples:
                generated_samples[key].extend(samples)
                continue

            # Not able to generate by faker so, generating samples by llm
            sampled_examples = random.sample(examples, min(3, len(examples)))

            response = self.llm_completion(
                prompt=tag_value_prompt.format(
                    num_samples_per_tag=num_samples_per_tag,
                    tag=tag,
                    tag_example=sampled_examples,
                )
            )
            generated_samples[key].extend(response.split("\n"))

        sampled_tags = random.choices(tags, k=min(10, len(tags)))

        templatized_sentences_examples = self.llm_completion(
            prompt=template_prompt.format(
                tags = ', '.join(sampled_tags).replace("'", '')
            ),
        )

        tasks = []
        for _ in range(num_call_batches):
            subsampled_dict = subsample_dictionary(attribute_values)

            sampled_keys = random.sample([*subsampled_dict], k = min(10, len(subsampled_dict)))

            values_requirements = "Take inspiration from the ideas below but do not mimic them directly. Ensure your output revolves around similar topics with some variations for accuracy..\n"
            for key in sampled_keys:
                values = subsampled_dict[key]
                values_requirements += (
                    f"Include the following {key}: {' and '.join(values)}.\n"
                )

            sampled_labels = random.sample([*tag_examples], k = min(5, len(tag_examples)))

            subsampled_label_sample_dict = subsample_dictionary(generated_samples, 5)
            label_sample_prompt = "Make sure to consider the examples listed below as inspiration. Your task is to generate entities that are similar in nature but distinctly unique\n"
            for key, values in subsampled_label_sample_dict.items():
                label_sample_prompt += f"{key}: {values}\n"

            sampled_random_prompts = [
                random.choices(items["prompts"], weights=items["scores"], k=1)[0]
                for __annotations__, items in self.random_prompts.items()
            ]
            rnd_prompts_str = "\n -\t".join(sampled_random_prompts)

            dataset_generation_prompt = f"""You possess deep expertise in {domain_prompt}. Please generate {batch_size} templates of synthetic sentences and associated tags for {domain_prompt}
            
            VERY IMPORTANT: MAKE SURE identify all named entities occurred that belong to one of the following entity types: 
            {sampled_labels}

            
            When generating the output for the NER task, adhere to the following strict format guidelines to ensure data consistency and format correctness
            
            Following are some sample output format for generation. This is just for example and you should not mimic this pattern.
            
            {templatized_sentences}

            Key Requirements:
            -   Mask the Only the Entities in square brackets with and make sure entities are in upper case. 
            -   The entities should strictly belong to one of {sampled_labels}. Do not include anything apart from entities in square brackets
            -   Seperate different samples by new line
            -   Give only the generated samples in output and make sure each sample should start on a new line. Do not include any extra new line. 
            -   DO NOT include any bulleting or prefix at start of each sentence and each. Do not include any quotes or emojis
            -   Give equal weightage to all the tags
            -   {rnd_prompts_str}
            
            {values_requirements}
            """

            tasks.append(dataset_generation_prompt)

        training_file_path = os.path.join(self.save_dir, "train.csv")

        random.shuffle(tasks)
        tasks = tasks[: total_expected_sentences - sentences_generated]

        error_logs = set()

        for task_batch_id in range(0, len(tasks), 20):
            task_batch = tasks[task_batch_id : task_batch_id + 20]
            results = []
            for task in tqdm(task_batch):
                results.append(process_task(task))

            for response, error_msg in results:
                if error_msg:
                    error_logs.add(error_msg)
                else:
                    try:
                        generated_templates = response.split("\n")
                        to_write_dataset = fill_and_transform_templates(
                            tags, generated_templates, generated_samples
                        )
                        file_mode = "w" if sentences_generated == 0 else "a"
                        sentences_generated += load_and_write_csv(
                            allowed_tags=tags + ["O"],
                            data_strings=to_write_dataset,
                            filename=training_file_path,
                            file_mode=file_mode,
                            fieldnames=["source", "target"],
                        )
                    except Exception as e:
                        print(f"Error writing data: {e}")

        dataset_config = {
            "filepath": training_file_path,
            "task": "TOKEN_CLASSIFICATION",
            "input_feature": "source",
            "target_feature": "target",
            "target_labels": tags,
            "num_samples": sentences_generated,
        }
        with open(os.path.join(self.save_dir, "config.json"), "w") as file:
            json.dump(dataset_config, file, indent=4)

        return dataset_config
