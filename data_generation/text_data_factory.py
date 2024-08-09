import random
import time
import traceback
from concurrent.futures import ProcessPoolExecutor, as_completed
from resource.text_prompts import datagen_prompt
from typing import Dict, List, Optional

from data_factory_interface import DataFactory
from tqdm import tqdm
from utils import assert_sufficient_descriptions, assert_sufficient_examples


class TextDataFactory(DataFactory):
    SOURCE_COLUMN = "text"
    TARGET_COLUMN = "label"

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
        vocab_per_sentence=4,
        sentences_generated=0,  # To resume the generate function incase of midway failure. TODO(Gautam): Incorporate resuming the data_generation task
    ):
        total_expected_sentences = samples_per_label * len(target_labels)
        assert sentences_generated < total_expected_sentences

        assert_sufficient_examples(target_labels, examples)
        assert_sufficient_descriptions(target_labels, labels_description)

        arguments = []

        generate_at_a_time = 40  # Max data-points to generate at a time
        for target_label in target_labels:
            for current_sentence_idx in range(0, samples_per_label, generate_at_a_time):
                samples_to_generate = min(
                    generate_at_a_time, samples_per_label - current_sentence_idx
                )
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
                    label_to_generate=target_label,
                    label_description=labels_description[target_label],
                    examples=label_examples,
                    user_prompts=(
                        ("\n".join(user_prompts) + "\n\n") if user_prompts else ""
                    ),
                    random_prompts="\n".join(self.get_random_prompts()),
                    random_vocab=str(random_vocab),
                )

                arguments.append({"prompt": prompt, "target_label": target_label})

        random.shuffle(arguments)
        arguments = arguments[: total_expected_sentences - sentences_generated]
        write_chunk_size = 50

        total_chunks = len(arguments) // write_chunk_size + 1
        for idx in range(0, len(arguments), write_chunk_size):
            chunk = arguments[idx : idx + write_chunk_size]
            data_points = []

            with ProcessPoolExecutor() as executor, tqdm(
                total=len(chunk),
                desc=f"Generating text data {(idx // write_chunk_size)}/{total_chunks}",
            ) as pbar:
                futures = []

                # Submit arguments to the executor
                for task in chunk:
                    future = executor.submit(
                        self.process_prompt,
                        task["prompt"],
                        task["target_label"],
                        task.get("system_prompt"),
                    )
                    future.add_done_callback(lambda p: pbar.update())
                    futures.append(future)

                # Wait for all arguments to complete and handle exceptions
                for future in as_completed(futures):
                    try:
                        response_text, target_label = future.result()
                        data_points.append(
                            {"texts": response_text, "target_label": target_label}
                        )

                    except Exception as e:
                        with open(self.errored_file_location, mode="a") as errored_fp:
                            traceback.print_exc(file=errored_fp)
                            errored_fp.write("\n" + "=" * 100 + "\n")

            transformed_data_points = []
            for data_point in data_points:
                temp = self.fill_and_transform(**data_point)
                if temp:
                    transformed_data_points.extend(temp)

            random.shuffle(transformed_data_points)

            self.write_on_training_file(
                transformed_data_points,
                fieldnames=[
                    TextDataFactory.SOURCE_COLUMN,
                    TextDataFactory.TARGET_COLUMN,
                ],
                write_fields=sentences_generated == 0,
                newline="",
                encoding="utf-8",
            )

            sentences_generated += len(transformed_data_points)

        dataset_config = {
            "filepath": str(self.train_file_location),
            "error_file": str(self.errored_file_location),
            "task": "TEXT_CLASSIFICATION",
            "input_feature": TextDataFactory.SOURCE_COLUMN,
            "target_feature": TextDataFactory.TARGET_COLUMN,
            "target_labels": target_labels,
            "num_samples": sentences_generated,
        }
        self.save_dict(self.config_file_location, **dataset_config)

        return dataset_config

    def fill_and_transform(self, texts: str, target_label: str):
        return [
            {
                TextDataFactory.SOURCE_COLUMN: text,
                TextDataFactory.TARGET_COLUMN: target_label,
            }
            for text in texts.replace("\n\n", "\n").split("\n")
            if text.strip()
        ]

    def process_prompt(
        self, prompt: str, target_label: str, system_prompt: Optional[str] = None
    ):
        texts_of = self.llm_completion(prompt=prompt, system_prompt=system_prompt)
        return texts_of, target_label
