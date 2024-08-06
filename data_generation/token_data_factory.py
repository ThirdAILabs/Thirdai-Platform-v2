import json
import os
import random
import traceback
from typing import List

from data_factory_interface import DataFactory
from faker import Faker
from tqdm import tqdm
from utils import (
    fill_and_transform_templates,
    get_faker_entities,
    load_and_write_csv,
    load_random_prompts,
    subsample_dictionary,
)


class TokenDataFactory(DataFactory):
    def __init__(
        self,
        api_key: str,
        random_prompts_file: str = "random_prompts.json",
    ):
        super().__init__(api_key)

        self.fake = Faker()
        self.random_prompts = load_random_prompts(random_prompts_file)
        providers = self.fake.providers
        self.faker_attributes = []

        provider_methods = {}

        for provider in providers:
            provider_name = provider.__class__.__name__
            methods = [method for method in dir(provider) if not method.startswith("_")]
            provider_methods[provider_name] = methods

        for provider_name, methods in provider_methods.items():
            for method in methods:
                self.faker_attributes.append(method)
        self.faker_attributes = sorted(self.faker_attributes, key=len)

    def generate(
        self,
        domain_prompt: str,
        tags: List[str],
        num_call_batches: int,
        tag_examples: dict,
        batch_size=40,
        num_samples_per_tag=100,
        sentences_generated=0,  # To resume the generate function incase of midway failure. TODO(Gautam): Incorporate resuming the data_generation task
    ):
        total_expected_sentences = batch_size * num_call_batches

        def process_task(prompt):
            try:
                response = self.openai.generate_output(
                    prompt,
                    system_prompt=f"You are a helpful assistant designed to generate synthetic data for domain {domain_prompt}.",
                )
                return response, None
            except Exception as e:
                print(f"An error occurred while generation: {str(e)}")
                print(traceback.format_exc())
                return None, error_msg

        assert all(tag in tags for tag in list(tag_examples.keys()))

        attribute_dimension_prompt = f"""Which attribute dimensions do you consider most vital in determining the topic of the task: ```{domain_prompt}```?

        DO NOT include any bulleting or prefix at start of each sentence and each. Do not include any quotes or emojis.

        VERY IMPORTANT POINT: Give only the attributes in output and make sure each attribute should start on a new line. Do not include any extra new line.

        List atmost 5 attributes.
        
        Eg. If domain is news, topics can be
        Sports
        Healthcare
        Politics
        Stocks
        Weather
        """

        response = self.openai.generate_output(
            prompt=attribute_dimension_prompt,
        )

        attributes = response.split("\n")

        print("Attributes:", attributes)

        attribute_values = {}
        for attribute in tqdm(attributes, desc="Attributed definition..."):
            attribute_value_prompt = f"""Given your extensive expertise in {domain_prompt}, please provide a range of realistic values for {attribute}. Ensure these estimates are diverse and applicable to real-world scenarios. 
            For attributes known to have well-defined values, provide specific, practical estimates; for all other attributes, offer generalized yet realistic ranges.

            DO NOT include any bulleting or prefix at start of each sentence and each. Do not include any quotes or emojis.

            VERY IMPORTANT POINT: Give only the attributes in output and make sure each attribute should start on a new line. Do not include any extra new line.
            """

            response = self.openai.generate_output(
                prompt=attribute_value_prompt,
            )

            attribute_values[attribute] = response.split("\n")

        generated_samples = {}
        for tag in tags:
            generated_samples[tag] = []

        ## Generate sample for entities
        for key, value in tqdm(
            tag_examples.items(), desc="Generating Sample for attributes.."
        ):
            samples, status = get_faker_entities(
                key,
                self.fake,
                self.faker_attributes,
                num_samples=total_expected_sentences,
            )
            if status:
                print(f"Sampled values for {key} from Faker")
                generated_samples[key].extend(samples)
                continue
            print(f"Generating Samples for {key}")
            subsampled_values = value
            if len(value) > 10:
                subsampled_values = random.sample(value, 10)
            generate_random_entities_prompt = f"""You possess deep expertise in {domain_prompt}. Please generate {num_samples_per_tag} diverse samples for the {key} named entity. Below are some examples of the {key} entity:
        {subsampled_values}

        Ensure each sample starts on a new line without any bullet points, prefixes, quotes, or emojis at the beginning of each sentence.

        VERY IMPORTANT: Only include the attributes themselves in the output. Each attribute should appear on a new line without any additional new lines.

        Additionally, aim to cover a wide range of variations within the {key} entity to ensure the data is as varied and representative as possible.
        """
            response = self.openai.generate_output(
                prompt=generate_random_entities_prompt,
            )
            generated_samples[key].extend(response.split("\n"))

        sampled_tags = random.choices(tags, k=min(3, len(tags)))

        generate_template_prompt = f"""You have to generate 2 templatized sentences for the tags: {str(sampled_tags).replace("'",'')}

        As an example here are two sentences for the tags [GENDER,ENTHNICITY,DISABILITY,SSN]

        She is [GENDER] and is of [ENTHNICITY] DESCENT 
        After getting diagnosed with [DISABILITY] John went home. His social security number is [SSN].

        Each sentence should start on a new line and with no bulleting or prefixing.
        """

        templatized_sentences = self.openai.generate_output(
            prompt=generate_template_prompt,
        )

        tasks = []
        for _ in range(num_call_batches):
            subsampled_dict = subsample_dictionary(attribute_values)

            sampled_keys = (
                random.sample(list(subsampled_dict.keys()), 10)
                if len(subsampled_dict) > 10
                else list(subsampled_dict.keys())
            )

            values_requirements = "Take inspiration from the ideas below but do not mimic them directly. Ensure your output revolves around similar topics with some variations for accuracy..\n"
            for key in sampled_keys:
                values = subsampled_dict[key]
                values_requirements += (
                    f"Include the following {key}: {' and '.join(values)}.\n"
                )

            sampled_labels = (
                random.sample(list(tag_examples.keys()), 5)
                if len(tag_examples) > 5
                else list(tag_examples.keys())
            )

            subsampled_label_sample_dict = subsample_dictionary(generated_samples, 5)
            label_sample_prompt = "Make sure to consider the examples listed below as inspiration. Your task is to generate entities that are similar in nature but distinctly unique\n"
            for key in subsampled_label_sample_dict:
                values = subsampled_label_sample_dict[key]
                label_sample_prompt += f"{key}: {values}\n"

            rnd_prompts = [
                random.choices(items["prompts"], weights=items["scores"], k=1)[0]
                for __annotations__, items in self.random_prompts.items()
            ]
            rnd_prompts_str = "\n -\t".join(rnd_prompts)

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
