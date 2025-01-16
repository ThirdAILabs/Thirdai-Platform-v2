import random
from typing import Any, List


def reservoir_sampling(
    candidates: List[Any],
    reservoir_size: int,
    current_size: int,
    total_items_seen: int,
    recency_multipler: float,
) -> List[Any]:

    assert reservoir_size > 0
    assert current_size >= 0
    assert total_items_seen >= 0
    assert recency_multipler > 0

    if len(candidates) == 0:
        return []

    if reservoir_size - current_size > len(candidates):
        return candidates

    items_to_add = []
    for candidate in candidates:
        total_items_seen += 1
        # Reservoir full, decide whether to replace an existing item
        probability = recency_multipler * (reservoir_size / (total_items_seen + 1))
        if random.random() <= probability:
            items_to_add.append(candidate)
        # Else, discard the candidate

    # if we end up selecting more items than the reservoir size, subsample
    element_to_sample = min(reservoir_size, len(items_to_add))
    items_to_add = random.sample(items_to_add, element_to_sample)

    return items_to_add
