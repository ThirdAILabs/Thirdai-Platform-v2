import csv
import json
import random
from typing import List, Dict


def subsample_dictionary(data: Dict[str, List[str]], k=2):
    return {
        key: random.sample(values, min(k, len(values))) for key, values in data.items()
    }


def convert_template_to_json(allowed_tags, data_strings):
    data = []
    for data_string in data_strings:
        if len(data_string.split()):
            try:
                words = data_string.split()
                tag_list = ["O"] * len(words)
                for i, word in enumerate(words):
                    if "[" in word and "]" in word:
                        start = word.index("[")
                        end = word.index("]")
                        assert word[start + 1 : end].upper() in allowed_tags
                        tag_list[i] = word[start + 1 : end].upper()
                json_obj = {
                    "source": words,
                    "target": tag_list,
                }
                data.append(json_obj)
            except Exception as e:
                print(f"Error processing data: {e}")
    return data


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
                    writer.writerow({"source": " ".join(src), "target": " ".join(tar)})
                    num_sentences += 1
            except Exception as e:
                print(f"Error processing data: {e}")
                return num_sentences
    return num_sentences
