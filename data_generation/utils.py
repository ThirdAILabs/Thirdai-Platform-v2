import csv
import json
import random
import re
from typing import Dict, List, Optional


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


def write_to_csv(path: str, data_points: List[str], header: List[str]):
    with open(path, "w") as csv_file:
        csv_writer = csv.DictWriter(csv_file, fieldnames=header)
        csv_writer.writeheader()
        csv_writer.writerows(data_points)
