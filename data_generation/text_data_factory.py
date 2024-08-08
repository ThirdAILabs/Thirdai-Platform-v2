import csv
import json
import os
import random
import traceback
from concurrent.futures import ProcessPoolExecutor, as_completed
from resource.text_prompts import datagen_prompt
from typing import Dict, List, Optional

from data_factory_interface import DataFactory
from tqdm import tqdm

from .utils import assert_sufficient_descriptions, assert_sufficient_examples

SOURCE_COLUMN = "text"
TARGET_COLUMN = "label"


class TextDataFactory(DataFactory):
    def __init__(self):
        super().__init__()

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

        assert_sufficient_examples(target_labels, examples)
        assert_sufficient_descriptions(target_labels, labels_description)

        def process_task(prompt: str, target_label: str):
            text_response = self.llm_model.completion(
                prompt,
            )
            return [
                {SOURCE_COLUMN: text, TARGET_COLUMN: target_label}
                for text in text_response.replace("\n\n", "\n").split("\n")
                if text.strip()
            ]

        input_data = []

        for target_label in target_labels:
            for batch_offset in range(0, samples_per_label, batch_size):
                samples_to_generate = min(batch_size, samples_per_label - batch_offset)
                random_vocab = self.get_random_vocab(
                    user_vocab, k=vocab_per_sentence * samples_to_generate
                )

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
                    user_prompts=(
                        ("\n".join(user_prompts) + "\n\n") if user_prompts else ""
                    ),
                    random_prompts="\n".join(self.get_random_prompts()),
                    random_vocab=str(random_vocab),
                )

                input_data.append((prompt, target_label))

        random.shuffle(input_data)
        input_data = input_data[: total_expected_sentences - sentences_generated]

        write_chunk_size = 50
        total_chunks = len(input_data) // write_chunk_size + 1
        for idx in range(0, len(input_data), write_chunk_size):
            chunk = input_data[idx : idx + write_chunk_size]
            data_points = []
            with ProcessPoolExecutor() as executor, tqdm(
                total=len(chunk),
                desc=f"Generating text data {(idx // write_chunk_size)}/{total_chunks}",
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
                            data_points.append(response)

                    except Exception as e:
                        with open(self.errored_file_location, mode="a") as errored_fp:
                            traceback.print_exc(file=errored_fp)
                            errored_fp.write("\n" + "=" * 100 + "\n")

            random.shuffle(data_points)

            self.write_on_training_file(
                data_points,
                fieldnames=(
                    [SOURCE_COLUMN, TARGET_COLUMN] if sentences_generated == 0 else None
                ),
                newline="",
                encoding="utf-8",
            )

            sentences_generated += len(data_points)

        dataset_config = {
            "filepath": self.train_file_location,
            "error_file": self.errored_file_location,
            "task": "TEXT_CLASSIFICATION",
            "input_feature": SOURCE_COLUMN,
            "target_feature": TARGET_COLUMN,
            "target_labels": target_labels,
            "num_samples": sentences_generated,
        }
        self.save_config(dataset_config)

        return dataset_config
