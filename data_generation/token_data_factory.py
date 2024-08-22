import random
import re
from collections import defaultdict
from resource.token_prompts import (
    dataset_generation_prompt,
    tag_value_prompt,
    template_prompt,
)
from typing import Dict, List

from data_factory_interface import DataFactory
from faker import Faker
from tqdm import tqdm
from utils import consistent_split, save_dict
from variables import Entity


class TokenDataFactory(DataFactory):
    SOURCE_COLUMN = "source"
    TARGET_COLUMN = "target"

    def __init__(
        self,
    ):
        super().__init__()
        self.faker = Faker()
        self.sentences_per_template = 4

        # All methods present in the faker to generate forged tags. E.g: credit_card_expire(), credit_card_expire(), first_name(), language_name(), ..
        self.faked_methods = [
            method
            for provider in self.faker.providers
            for method in dir(provider)
            if not method.startswith("_")
        ]

    def get_fake_tag_values(self, tag: str, num_samples: int):
        # NOTE: It could be better to have an exact match
        matched_attrs = list(
            filter(lambda method: tag.lower() in method.lower(), self.faked_methods)
        )
        if not matched_attrs:
            return []

        matched_attr = min(matched_attrs, key=len)

        return list(
            map(
                lambda x: str(x),
                [self.faker.__getattr__(matched_attr)() for _ in range(num_samples)],
            ),
        )

    def get_complete_tag_examples(
        self,
        domain_prompt: str,
        tags: List[Entity],
        total_expected_sentences: int,
        num_samples_per_tag: int,
    ) -> Dict[str, List[str]]:
        complete_tag_examples = defaultdict(list)

        for tag in tqdm(tags, desc="Generating examples for tags: ", leave=False):
            complete_tag_examples[tag.name].extend(tag.examples)

            # Trying to generate more examples from faker
            samples = self.get_fake_tag_values(
                tag.name,
                num_samples=total_expected_sentences,
            )
            if samples:
                complete_tag_examples[tag.name].extend(samples)
                continue

            # Not able to generate by faker so, generating samples by llm
            sampled_tag_examples = random.sample(
                tag.examples, k=min(3, len(tag.examples))
            )
            response = self.llm_model.completion(
                prompt=tag_value_prompt.format(
                    domain_prompt=domain_prompt,
                    num_samples_per_tag=num_samples_per_tag,
                    tag=tag.name,
                    tag_example=str(sampled_tag_examples),
                    tag_description=tag.description,
                )
            )
            complete_tag_examples[tag.name].extend(response.split("\n"))

        return complete_tag_examples

    def get_templatized_examples(self, tags: List[Entity], k: int = 2):
        return self.llm_model.completion(
            template_prompt.format(
                tags=", ".join([tag.name for tag in tags]),
                tags_description="\n".join(
                    [f"{tag.name}: {tag.description}" for tag in tags]
                ),
                k=k,
            )
        )

    def generate_data(
        self,
        domain_prompt: str,
        tags: List[Entity],
        num_sentences_to_generate: int,
        num_samples_per_tag: int = 4,
    ):
        complete_tag_examples = self.get_complete_tag_examples(
            domain_prompt, tags, num_sentences_to_generate, num_samples_per_tag
        )  # Here num_sentences_to_generate is used by faker to generate this many value of each tag.
        save_dict(self.save_dir / "tag_examples.json", **complete_tag_examples)

        templatized_sentences_examples = self.get_templatized_examples(
            tags=random.sample(tags, k=min(10, len(tags))), k=2
        )

        arguments = []
        templates_to_generate = (
            num_sentences_to_generate - self.sentences_generated
        ) // self.sentences_per_template
        for current_sentence_idx in range(
            0, templates_to_generate, self.generate_at_a_time
        ):
            # TODO(anyone): we should also add the [user_tag -> examples] in dataset_generation_prompt.
            sampled_tags = random.sample(tags, k=min(5, len(tags)))
            arguments.append(
                {
                    "prompt": dataset_generation_prompt.format(
                        domain_prompt=domain_prompt,
                        num_to_generate=min(
                            self.generate_at_a_time,
                            templates_to_generate - current_sentence_idx,
                        ),
                        tags=[t.name for t in sampled_tags],
                        tag_description="\n".join(
                            [f"{t.name}: {t.description}" for t in sampled_tags]
                        ),
                        templatized_sentences_examples=templatized_sentences_examples,
                        rnd_prompts_str="\n-  ".join(self.get_random_prompts()),
                    ),
                    "system_prompt": f"You are a helpful assistant designed to generate synthetic data for domain {domain_prompt}.",
                }
            )

        random.shuffle(arguments)

        total_chunks = len(arguments) // self.write_chunk_size + 1
        for idx in tqdm(
            range(0, len(arguments), self.write_chunk_size),
            desc="Generating token data: ",
            total=total_chunks,
        ):
            chunk_to_process = arguments[idx : idx + self.write_chunk_size]

            generated_templates: List[str] = self.run_and_collect_results(
                tasks_prompt=chunk_to_process, parallelize=True
            )

            transformed_data_points = [
                data_point
                for template_s in generated_templates
                for template in template_s["response_text"].split("\n")
                for data_point in self.fill_and_transform(
                    template, complete_tag_examples
                )
            ]
            # filtering to remove 'None'
            transformed_data_points = list(
                filter(lambda x: x not in [None, [], {}], transformed_data_points)
            )

            random.shuffle(transformed_data_points)

            self.write_on_training_file(
                transformed_data_points,
                fieldnames=[
                    TokenDataFactory.SOURCE_COLUMN,
                    TokenDataFactory.TARGET_COLUMN,
                ],
            )

            self.sentences_generated += len(transformed_data_points)

        dataset_config = {
            "filepath": str(self.train_file_location),
            "error_file": str(self.errored_file_location),
            "task": "TOKEN_CLASSIFICATION",
            "input_feature": TokenDataFactory.SOURCE_COLUMN,
            "target_feature": TokenDataFactory.TARGET_COLUMN,
            "target_labels": [tag.name for tag in tags],
            "num_samples": self.sentences_generated,
        }
        save_dict(self.config_file_location, **dataset_config)

        return dataset_config

    def fill_and_transform(
        self, template: str, tag_values: Dict[str, List[str]]
    ) -> List[str]:
        if not template:
            return [None]

        seperator = " "
        words = consistent_split(template, seperator)

        data_points = [
            {TokenDataFactory.SOURCE_COLUMN: [], TokenDataFactory.TARGET_COLUMN: []}
            for _ in range(self.sentences_per_template)
        ]

        for word in words:
            match = re.search(r"\[(.*?)\]", word)
            if match:
                # word is a tag
                word_tag = match.group(1).upper()
                if word_tag not in tag_values:
                    self.write_on_errorfile(
                        text=f"\nTag {word_tag} not present in the allowed tags {', '.join(tag_values.keys())}"
                        + f"\ntemplate: {template}"
                    )
                    return [None]

                for idx in range(self.sentences_per_template):
                    splitted_word_tag_value = consistent_split(
                        random.choice(tag_values[word_tag]), seperator
                    )
                    data_points[idx][TokenDataFactory.SOURCE_COLUMN].extend(
                        splitted_word_tag_value
                    )

                    """
                    Extending the [TAG] to match the source text

                        For example:
                            template = '[NAME] reserved the hall for reunion'
                            word_tag = [NAME]
                            word_tag_value = Jessica vega
                            splitted_word_tag_value = ['Jessica', 'vega']

                        Expected:
                            source = 'Jessica vega reserved the hall for reunion'
                            target = 'NAME NAME O O O O O'
                    """
                    data_points[idx][TokenDataFactory.TARGET_COLUMN].extend(
                        [word_tag] * len(splitted_word_tag_value)
                    )
            else:
                for idx in range(self.sentences_per_template):
                    data_points[idx][TokenDataFactory.SOURCE_COLUMN].append(word)
                    data_points[idx][TokenDataFactory.TARGET_COLUMN].append("O")

        # make sure that the source and target is of same length
        for i in range(self.sentences_per_template):
            data = data_points[i]
            if len(data[TokenDataFactory.SOURCE_COLUMN]) != len(
                data[TokenDataFactory.TARGET_COLUMN]
            ):
                self.write_on_errorfile(
                    text=f"\nSource and target aren't of the same length {'-' * 30}"
                    + f"\nsource: {data[TokenDataFactory.SOURCE_COLUMN]}"
                    + f"\ntarget: {data[TokenDataFactory.TARGET_COLUMN]}"
                    + f"\n\ntemplate: {template}"
                )
                data_points.pop(i)
            else:
                data[TokenDataFactory.SOURCE_COLUMN] = seperator.join(
                    data[TokenDataFactory.SOURCE_COLUMN]
                )
                data[TokenDataFactory.TARGET_COLUMN] = seperator.join(
                    data[TokenDataFactory.TARGET_COLUMN]
                )

        return data_points
