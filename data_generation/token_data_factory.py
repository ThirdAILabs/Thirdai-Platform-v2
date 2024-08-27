import math
import random
import re
from collections import defaultdict
from resource.token_prompts import (
    dataset_generation_prompt,
    tag_value_prompt,
    template_prompt,
)
from typing import Dict, List, Optional

from data_factory_interface import DataFactory
from faker import Faker
from tqdm import tqdm
from utils import consistent_split, remove_duplicates, save_dict, write_to_csv
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
        self.max_value_per_tag_to_generate = 1000

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
            filter(lambda method: tag.lower() == method.lower(), self.faked_methods)
        )

        if not matched_attrs:
            return []

        return list(
            map(
                lambda x: str(x),
                [
                    self.faker.__getattr__(matched_attrs[0])()
                    for _ in range(num_samples)
                ],
            ),
        )

    def get_complete_tag_examples(
        self,
        domain_prompt: str,
        tags: List[Entity],
        num_examples_to_generate: int,
        examples_to_generate_at_a_time: int = 100,
    ) -> Dict[str, List[str]]:
        complete_tag_examples = defaultdict(list)

        for tag in tqdm(tags, desc="Generating examples for tags: ", leave=False):
            complete_tag_examples[tag.name].extend(tag.examples)

            # Trying to generate more examples from faker
            samples = self.get_fake_tag_values(
                tag.name,
                num_samples=num_examples_to_generate,
            )
            if samples:
                complete_tag_examples[tag.name].extend(samples)
                continue

            # Not able to generate by faker so, generating samples by llm
            sampled_tag_examples = random.sample(
                tag.examples, k=min(3, len(tag.examples))
            )
            value_per_tag_to_generate = min(
                self.max_value_per_tag_to_generate, num_examples_to_generate
            )
            for idx in range(
                0, value_per_tag_to_generate, examples_to_generate_at_a_time
            ):
                response = self.llm_model.completion(
                    prompt=tag_value_prompt.format(
                        domain_prompt=domain_prompt,
                        num_samples_per_tag=min(
                            examples_to_generate_at_a_time,
                            value_per_tag_to_generate - idx,
                        ),
                        tag=tag.name,
                        tag_example=str(sampled_tag_examples),
                        tag_description=tag.description,
                    )
                )
            complete_tag_examples[tag.name].extend(
                remove_duplicates(response.split("\n"))
            )

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

    def train_test_tag_split(
        self,
        tags_example: Dict[str, List[str]],
        shuffle: bool = True,
        save: bool = True,
    ):
        if shuffle:
            tags_example = {
                key: random.shuffle(value) for key, value in tags_example.items()
            }

        if not self.general_variables.test_size:
            if save:
                save_dict(self.save_dir / "train", "tag_value.json", **tags_example)

            return tags_example, None

        train_tags, test_tags = {}, {}
        for tag_name, examples in tags_example.items():
            split_index = int(len(examples) * (1 - self.general_variables.test_size))
            train_tags[tag_name] = examples[:split_index]
            test_tags[tag_name] = examples[split_index:]

        if save:
            save_dict(self.save_dir / "train", "tag_value.json", **train_tags)
            save_dict(self.save_dir / "test", "tag_value.json", **test_tags)

        return train_tags, test_tags

    def train_test_template_split(
        self, templates: List[str], shuffle: bool = True, save: bool = True
    ):
        templates = list(filter(lambda x: x not in [None, [], {}, ""], templates))

        if shuffle:
            random.shuffle(templates)

        if not self.general_variables.test_size:
            if save:
                write_to_csv(
                    path=self.save_dir / "train" / "templates.csv",
                    data_points=templates,
                    header=["template"],
                )

            return templates, None

        split_index = int(len(templates) * (1 - self.general_variables.test_size))
        train_templates, test_templates = (
            templates[:split_index],
            templates[split_index:],
        )
        if save:
            write_to_csv(
                path=self.save_dir / "train" / "templates.csv",
                header=["template"],
                texts=train_templates,
            )
            write_to_csv(
                path=self.save_dir / "test" / "templates.csv",
                header=["template"],
                texts=test_templates,
            )
        return train_templates, test_templates

    def _templates_to_generate(self, num_sentences_to_generate: int):
        return (
            num_sentences_to_generate - self.train_sentences_generated
        ) // self.sentences_per_template

    def _subsample_tag(self, tags: List[Entity]):
        # using triangular distribution to favour longer lists by setting mode = high.
        k = math.ceil(random.triangular(low=1, high=len(tags), mode=len(tags)))
        return random.sample(tags, k)

    def generate_data(
        self,
        domain_prompt: str,
        tags: List[Entity],
        num_sentences_to_generate: int,
        examples_per_tag_to_generate: Optional[int] = None,
    ):
        templates_to_generate = self._templates_to_generate(num_sentences_to_generate)

        complete_tag_examples = self.get_complete_tag_examples(
            domain_prompt=domain_prompt,
            tags=tags,
            num_examples_to_generate=examples_per_tag_to_generate
            or (templates_to_generate * self.sentences_per_template),
        )

        train_tag_examples, test_tag_examples = self.train_test_tag_split(
            tags_example=complete_tag_examples, shuffle=True, save=True
        )

        templatized_sentences_examples = self.get_templatized_examples(
            tags=random.sample(tags, k=min(10, len(tags))), k=2
        )

        arguments = []
        for current_sentence_idx in range(
            0, templates_to_generate, self.generate_at_a_time
        ):
            # TODO(anyone): we should also add the [user_tag -> examples] in dataset_generation_prompt.
            sampled_tags = self._subsample_tag(tags)

            arguments.append(
                {
                    "prompt": dataset_generation_prompt.format(
                        domain_prompt=domain_prompt,
                        num_to_generate=min(
                            templates_to_generate - current_sentence_idx,
                            self.generate_at_a_time,
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

            templates = [
                template
                for template_s in generated_templates
                for template in template_s["response_text"].split("\n")
            ]

            train_templates, test_templates = self.train_test_template_split(
                templates=templates, shuffle=True, save=True
            )
            train_datapoints = [
                self.fill_and_transform(
                    template=template, tag_values=train_tag_examples
                )
                for template in train_templates
            ]
            random.shuffle(train_datapoints)
            train_datapoints = list(
                filter(lambda x: x not in [None, [], {}, ""], train_datapoints)
            )

            if test_templates:
                test_datapoints = [
                    self.fill_and_transform(
                        template=template, tag_values=test_tag_examples
                    )
                    for template in test_templates
                ]
                random.shuffle(test_datapoints)
                test_datapoints = list(
                    filter(lambda x: x not in [None, [], {}, ""], test_datapoints)
                )

            self.write_on_file(
                train_datapoints,
                fieldnames=[
                    TokenDataFactory.SOURCE_COLUMN,
                    TokenDataFactory.TARGET_COLUMN,
                ],
                is_train_file=True,
            )
            self.train_sentences_generated += len(train_datapoints)

            if self.general_variables.test_size:
                self.write_on_file(
                    test_datapoints,
                    fieldnames=[
                        TokenDataFactory.SOURCE_COLUMN,
                        TokenDataFactory.TARGET_COLUMN,
                    ],
                    is_train_file=False,
                )
                self.test_sentences_generated += len(test_datapoints)

        dataset_config = {
            "filepath": str(self.train_file_location),
            "error_file": str(self.errored_file_location),
            "task": "TOKEN_CLASSIFICATION",
            "input_feature": TokenDataFactory.SOURCE_COLUMN,
            "target_feature": TokenDataFactory.TARGET_COLUMN,
            "target_labels": [tag.name for tag in tags],
            "train_samples": self.train_sentences_generated,
            "test_samples": self.test_sentences_generated,
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
        for i in range(len(data_points)):
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
