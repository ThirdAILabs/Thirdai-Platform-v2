import json
import os
from collections import defaultdict, deque
from datetime import datetime
from typing import Union

from deployment_job.pydantic_models.inputs import AssociateInput, UpvoteInput


class FeedbackCollector:
    def __init__(self, log_dir, track_last_n: int = 20, write_after_updates: int = 5):
        os.makedirs(log_dir, exist_ok=True)
        self._log_file = os.path.join(log_dir, f"{os.getenv('NOMAD_ALLOC_ID')}.json")
        self._queue = defaultdict(lambda: deque(maxlen=track_last_n))
        self.write_after_updates = write_after_updates
        self.update_counter = 0

    def add(self, input: Union[AssociateInput, UpvoteInput]):
        current_time = str(datetime.now().strftime("%d %B %Y %H:%M:%S"))
        if isinstance(input, AssociateInput):
            for text_id_pair in input.text_id_pairs:
                self._queue["associate"].append(
                    {
                        "timestamp": current_time,
                        "query_text": text_id_pair.query_text,
                        "reference_id": text_id_pair.reference_id,
                        "reference_text": text_id_pair.reference_text,
                    }
                )
        elif isinstance(input, UpvoteInput):
            for text_pair in input.text_pairs:
                self._queue["upvote"].append(
                    {
                        "timestamp": current_time,
                        "source": text_pair.source,
                        "target": text_pair.target,
                    }
                )
        else:
            raise ValueError("Input type not supported for logging")

        self.update_counter += 1

        if self.update_counter % self.write_after_updates == 0:
            queue_to_write = {
                event: list(items) for event, items in self._queue.items()
            }

            # write updates to jsonl file
            with open(self._log_file, "w") as fp:
                json.dump(queue_to_write, fp, indent=4)

            # reset update counter
            self.update_counter = 0

    @staticmethod
    def get_feedback_logger(deployment_dir: str):
        return FeedbackCollector(os.path.join(deployment_dir, "recent_feedbacks"))
