import csv
import json
import random
import re
from typing import Dict, List


def assert_sufficient_examples(
    target_labels: List[str], examples: Dict[str, List[str]]
):
    missing_examples = [label for label in target_labels if label not in examples]
    if missing_examples:
        raise ValueError(
            f"Examples are not given for all labels. Labels with missing examples: {', '.join(missing_examples)}"
        )


def assert_sufficient_descriptions(
    target_labels: List[str], labels_description: Dict[str, str]
):
    missing_description = [
        label for label in target_labels if label not in labels_description
    ]
    if missing_description:
        raise ValueError(
            f"Descriptions are not given for all labels. Labels with missing descriptions: {', '.join(missing_description)}"
        )


def subsample_dictionary(data: Dict[str, List[str]], k=2):
    return {
        key: random.sample(values, min(k, len(values))) for key, values in data.items()
    }


def parse_template(template: str, allowed_tags: List[str]):
    words = template.split()
    if not words:
        return None, None, None

    source_template = []
    words_tag = []
    tags_present = []
    for word in words:
        match = re.search(r"\[(.*?)\]", word)
        if match:
            # word is a tag
            word_tag = match.group(1)
            assert word_tag in allowed_tags

            words_tag.append(word_tag)
            source_template.append(word.replace(match.group(0), f"{{{word_tag}}}", 1))
            tags_present.append(word_tag)
        else:
            source_template.append(word)
            words_tag.append("O")

    return " ".join(source_template), " ".join(words_tag), tags_present


def fill_and_transform_templates(
    allowed_tags, templates: List[str], tag_values: Dict[str, List[str]]
):
    data_points = []
    for template in templates:
        source_template, words_tag, tags_present = parse_template(
            template, allowed_tags
        )
        if not source_template:
            continue

        source_text = source_template.format(
            **{tag: random.choice(tag_values[tag]) for tag in tags_present}
        )

        data_points.append({"source": source_text, "target": words_tag})

    return data_points
