import csv
import json
import os
import random
import traceback
from concurrent.futures import ProcessPoolExecutor, as_completed
from resource.util_data import vocab
from typing import Dict, List, Optional

from data_factory_interface import DataFactory
from tqdm import tqdm
from utils import datagen_prompt

TEXT_COLUMN = "text"
TARGET_LABEL_COLUMN = "label"


def assert_sufficient_examples(
    target_labels: List[str], examples: Dict[str, List[str]]
):
    missing_examples = [
        label for label in examples.keys() if label not in target_labels
    ]
    if missing_examples:
        raise ValueError(
            f"Examples are not given for all labels. Labels with missing examples: {', '.join(missing_examples)}"
        )


def assert_sufficient_descriptions(
    target_labels: List[str], labels_description: Dict[str, str]
):
    missing_description = [
        label for label in labels_description.keys() if label not in target_labels
    ]
    if missing_description:
        raise ValueError(
            f"Descriptions are not given for all labels. Labels with missing descriptions: {', '.join(missing_description)}"
        )


class TextDataFactory(DataFactory):
    def __init__(self, api_key: str):
        super().__init__(api_key)
        self.vocab = vocab

    def generate_data(
        self,
        task_prompt: str,
        samples_per_label: int,
        target_labels: List[str],
        examples: Dict[str, List[str]],
        labels_description: Dict[str, str],
        user_vocab: Optional[List[str]] = None,
        user_prompts: Optional[List[str]] = None,
        batch_size=40,
        vocab_per_sentence=4,
        sentences_generated=0,  # To resume the generate function incase of midway failure. TODO(Gautam): Incorporate resuming the data_generation task
    ):
        total_expected_sentences = samples_per_label * len(target_labels)
        assert sentences_generated < total_expected_sentences

        def process_task(prompt: str, target_label: str):
            text_response = self.llm_completion(
                prompt,
            )
            return [
                (text, target_label)
                for text in text_response.replace("\n\n", "\n").split("\n")
                if text.strip()
            ]

        assert_sufficient_examples(target_labels, examples)
        assert_sufficient_descriptions(target_labels, labels_description)

        vocabulary = self.vocab + (user_vocab if user_vocab is not None else [])
        user_prompts = ("\n".join(user_prompts) + "\n\n") if user_prompts else ""

        input_data = []

        for target_label in target_labels:
            for batch_offset in range(0, samples_per_label, batch_size):
                samples_to_generate = min(batch_size, samples_per_label - batch_offset)
                random_vocab = random.sample(
                    population=vocabulary, k=vocab_per_sentence * samples_to_generate
                )

                sampled_random_prompts = [
                    random.choices(items["prompts"], weights=items["scores"], k=1)[0]
                    for __annotations__, items in self.random_prompts.items()
                ]

                label_examples = "\n".join(
                    random.sample(
                        examples[target_label],
                        min(2, len(examples[target_label])),
                    )
                )

                prompt = datagen_prompt.format(
                    task_prompt=task_prompt,
                    samples_to_generate=samples_to_generate,
                    target_label=target_label,
                    label_description=labels_description[target_label],
                    examples=label_examples,
                    user_prompts=user_prompts,
                    random_prompts="\n".join(sampled_random_prompts),
                    random_vocab=str(random_vocab),
                )

                input_data.append((prompt, target_label))

        random.shuffle(input_data)
        input_data = input_data[: total_expected_sentences - sentences_generated]
        file_mode = "w" if sentences_generated == 0 else "a"
        train_file_location = os.path.join(self.save_dir, "train.csv")
        errored_file_location = os.path.join(self.save_dir, "traceback.err")

        with open(
            train_file_location, file_mode, newline="", encoding="utf-8"
        ) as csvfile:
            csv_writer = csv.writer(csvfile)
            if sentences_generated == 0:
                csv_writer.writerow([TEXT_COLUMN, TARGET_LABEL_COLUMN])

            write_chunk_size = 20
            total_chunks = len(input_data) // write_chunk_size + 1
            for idx in range(0, len(input_data), write_chunk_size):
                chunk = input_data[idx : idx + write_chunk_size]
                with ProcessPoolExecutor() as executor, tqdm(
                    total=len(chunk),
                    desc=f"Generating text data {(idx // write_chunk_size)}/{total_chunks}",
                ) as pbar:
                    futures = []
                    text_with_target_label = []

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
                                text_with_target_label.append(response)

                        except Exception as e:
                            with open(errored_file_location, mode="a") as errored_fp:
                                traceback.print_exc(file=errored_fp)
                                errored_fp.write("\n" + "=" * 100 + "\n")

                random.shuffle(text_with_target_label)

                # Writing to the csv
                csv_writer.writerows(text_with_target_label)
                sentences_generated += len(text_with_target_label)

        dataset_config = {
            "filepath": train_file_location,
            "error_file": errored_file_location,
            "task": "TEXT_CLASSIFICATION",
            "input_feature": TEXT_COLUMN,
            "target_feature": TARGET_LABEL_COLUMN,
            "target_labels": target_labels,
            "num_samples": sentences_generated,
        }
        with open(os.path.join(self.save_dir, "config.json"), "w") as file:
            json.dump(dataset_config, file, indent=4)

        return dataset_config
