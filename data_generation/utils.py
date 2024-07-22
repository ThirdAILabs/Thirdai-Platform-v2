import csv
import json
import os
import random
from typing import List

datagen_prompt = """This is the description of the task user wants to perform: ```{task_prompt}```
Generate {samples_to_generate} training samples for above task for the label "{label_to_generate}".

{label_description_prompt}
DO NOT include any bulleting or prefix at start of each sentence and each. Do not include any quotes or emojis.


VERY IMPORTANT POINT: Give only the sentences in output and make sure each sentence should start on a new line. Do not include any extra new line.


Ensure that the data is diverse enough to capture different genres and dialects.
Do not include the label in sentences.

{example_prompt}{user_prompts_combined}
You can refer to these prompt to include diversity:
{rnd_prompts_str}

Sentences should have following words to generate diverse examples:
{random_vocab_str}
"""


def load_vocab(vocab_paths: List[str]):
    vocab = set()
    resource_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "resource")
    for vocab_path in vocab_paths:
        with open(os.path.join(resource_dir, vocab_path), "r") as file:
            for line in file:
                vocab.add(line.strip())
    return list(vocab)


def load_random_prompts(filepath: str):
    resource_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "resource")
    with open(os.path.join(resource_dir, filepath), "r") as file:
        data = json.load(file)
    prompts_json = {}
    for k, v in data.items():
        prompts_json[k] = {
            "prompts": [item["transformation"] for item in v],
            "scores": [item["score"] for item in v],
        }
    return prompts_json


def get_faker_entities(tag: str, fake, faker_attributes: List[str], num_samples: int):
    matching_attr = next(
        (attr for attr in faker_attributes if tag.lower() in attr.lower()), None
    )

    if not matching_attr:
        return [], False

    try:
        samples = [getattr(fake, matching_attr)() for _ in range(num_samples)]
        return samples, True
    except Exception as e:
        print(f"Error in faker {e}")
        return [], False


def subsample_dictionary(data, k=2):
    new_dict = {}
    for key, values in data.items():
        subsampled_values = random.sample(values, min(k, len(values)))
        new_dict[key] = subsampled_values
    return new_dict


def fill_and_transform_templates(
    allowed_tags, templates: List[str], generated_samples: dict
):
    raw_json_templates = convert_template_to_json(allowed_tags, templates)
    data_points = []
    for json_obj in raw_json_templates:

        source = json_obj["source"]
        target = json_obj["target"]

        new_source = []
        new_target = []
        for i in range(len(source)):
            if target[i] == "O":
                new_target.append("O")
                new_source.append(source[i])
            else:
                random_entity = random.choice(generated_samples[target[i]])
                random_entity = random_entity.split()
                assert random_entity != None
                new_target.extend([target[i]] * len(random_entity))
                new_source.extend(random_entity)
        new_json_obj = {"source": new_source, "target": new_target}
        if len(new_json_obj["source"]):
            data_points.append(new_json_obj)
    return data_points


def load_and_write_csv(allowed_tags, data_strings, filename, file_mode, fieldnames):
    num_sentences = 0
    with open(filename, file_mode) as csvfile:

        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)

        if file_mode == "w":
            writer.writeheader()

        for data_string in data_strings:
            try:
                if not isinstance(data_string, dict):
                    formatted_string = data_string.replace("'", '"')
                    json_object = json.loads(formatted_string)
                else:
                    json_object = data_string

                if all(tag in allowed_tags for tag in json_object["target"]):
                    src, tar = zip(
                        *[
                            (s, t)
                            for s, t in zip(
                                json_object["source"], json_object["target"]
                            )
                            if s != '"'
                        ]
                    )
                    # file.write(json.dumps(json_object) + "\n")
                    writer.writerow({"source": " ".join(src), "target": " ".join(tar)})
                    num_sentences += 1
            except Exception as e:
                print(f"Error processing data: {e}")
                return num_sentences
    return num_sentences
