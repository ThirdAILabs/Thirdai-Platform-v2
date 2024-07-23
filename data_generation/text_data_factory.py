import csv
import json
import os
import random
import traceback
from typing import Dict, List, Optional

from utils import datagen_prompt, load_random_prompts, load_vocab

from data_generation.data_factory_interface import DataFactory


class TextDataFactory(DataFactory):
    def __init__(
        self,
        api_key: str,
        random_prompts_file: str = "random_prompts.json",
        vocab: str = "general",
    ):
        super().__init__(api_key)
        vocab_paths = []
        if vocab == "general":
            general_vocab_path = "general_vocab.txt"
            vocab_paths.append(general_vocab_path)
        self.vocab = load_vocab(vocab_paths)
        self.random_prompts = load_random_prompts(random_prompts_file)

    def generate(
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

        def process_task(prompt):
            try:
                response = self.openai.generate_output(
                    prompt,
                )
                return response, None
            except Exception as e:
                print(f"An error occurred while generation: {str(e)}")
                print(traceback.format_exc())
                return None, error_msg

        assert all(label in target_labels for label in list(examples.keys()))

        user_vocab = self.vocab + (user_vocab if user_vocab is not None else [])
        tasks = []
        if user_prompts:
            user_prompts_combined = "\n".join(user_prompts)
            user_prompts_combined = f"{user_prompts_combined}\n\n"
        else:
            user_prompts_combined = ""

        for label_to_generate in target_labels:
            for batch_id in range(0, samples_per_label, batch_size):
                samples_to_generate = min(batch_size, samples_per_label - batch_id)
                random_vocab = random.sample(
                    user_vocab, vocab_per_sentence * samples_to_generate
                )

                rnd_prompts = [
                    random.choices(items["prompts"], weights=items["scores"], k=1)[0]
                    for __annotations__, items in self.random_prompts.items()
                ]
                rnd_prompts_str = "\n".join(rnd_prompts)

                if label_to_generate in examples:
                    examples_combined = "\n".join(
                        random.sample(
                            examples[label_to_generate],
                            min(2, len(examples[label_to_generate])),
                        )
                    )
                    example_prompt = (
                        f"Following are some of the sample data points for reference:\n{examples_combined}"
                        f"GENERATED SAMPLES MUST BE VERY DIFFERENT FROM THE ABOVE SAMPLES\n\n"
                    )
                else:
                    example_prompt = ""

                prompt = datagen_prompt.format(
                    task_prompt=task_prompt,
                    samples_to_generate=samples_to_generate,
                    label_to_generate=label_to_generate,
                    label_description_prompt=(
                        f"The data generated should strictly follow the description: {labels_description[label_to_generate]}\n\n"
                        if label_to_generate in labels_description
                        else ""
                    ),
                    example_prompt=example_prompt,
                    user_prompts_combined=user_prompts_combined,
                    rnd_prompts_str=rnd_prompts_str,
                    random_vocab_str=str(random_vocab),
                )

                tasks.append([prompt, label_to_generate])
        random.shuffle(tasks)
        tasks = tasks[: total_expected_sentences - sentences_generated]
        file_mode = "w" if sentences_generated == 0 else "a"
        train_file_location = os.path.join(self.save_dir, "train.csv")

        error_logs = set()
        with open(
            train_file_location, file_mode, newline="", encoding="utf-8"
        ) as csvfile:
            csv_writer = csv.writer(csvfile)
            if sentences_generated == 0:
                csv_writer.writerow(["text", "label"])

            for task_batch_id in range(0, len(tasks), 20):
                task_batch = tasks[task_batch_id : task_batch_id + 20]
                generated_results = {}

                results = [
                    (process_task(prompt), label_to_generate)
                    for (prompt, label_to_generate) in task_batch
                ]
                for (response, error_msg), label_to_generate in results:
                    if error_msg:
                        error_logs.add(error_msg)
                    else:
                        generated_result = generated_results.setdefault(
                            label_to_generate, ""
                        )
                        generated_results[label_to_generate] = (
                            generated_result + response.replace("\n\n", "\n") + "\n"
                        )

                data_list = []
                for label, texts in generated_results.items():
                    cleaned_texts = [text for text in texts.split("\n") if text.strip()]
                    data_list.extend((text, label) for text in cleaned_texts)
                sentences_generated += len(data_list)

                # throw error for testing
                random.shuffle(data_list)
                for text, label in data_list:
                    csv_writer.writerow([text, label])

        dataset_config = {
            "filepath": train_file_location,
            "task": "TEXT_CLASSIFICATION",
            "input_feature": "text",
            "target_feature": "label",
            "target_labels": target_labels,
            "num_samples": sentences_generated,
        }
        with open(os.path.join(self.save_dir, "config.json"), "w") as file:
            json.dump(dataset_config, file, indent=4)

        return dataset_config
