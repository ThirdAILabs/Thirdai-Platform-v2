import json
import re
from typing import List


def save_dict(write_to: str, **kwargs):
    with open(write_to, "w") as fp:
        json.dump(kwargs, fp, indent=4)


def consistent_split(text: str, seperator: str = " ") -> List[str]:
    return re.sub("\s+", seperator, text).strip().split(sep=seperator)


def remove_duplicates(words: List[str]):
    seen = set()
    uniques = []
    for item in words:
        if item.lower() not in seen:
            seen.add(item.lower())
            uniques.append(item)

    return uniques


def load_json(path: str):
    with open(path, "r") as fp:
        return json.load(fp)
