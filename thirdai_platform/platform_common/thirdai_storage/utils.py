import random
from typing import Any, List


def reservoir_sampling(
    candidates: List[Any],
    reservoir_size: int,
    current_size: int,
    total_items_seen: int,
    recency_multipler: float,
) -> List[Any]:
    items_to_add = []

    for candidate in candidates:
        total_items_seen += 1
        if current_size < reservoir_size:
            # Reservoir not full, add candidate
            items_to_add.append(candidate)
            current_size += 1
        else:
            # Reservoir full, decide whether to replace an existing item
            probability = recency_multipler * (
                reservoir_size / (total_items_seen + reservoir_size)
            )
            if random.random() <= probability:
                items_to_add.append(candidate)
            # Else, discard the candidate

    element_to_sample = min(reservoir_size, len(items_to_add))
    items_to_add = random.sample(items_to_add, element_to_sample)

    return items_to_add
