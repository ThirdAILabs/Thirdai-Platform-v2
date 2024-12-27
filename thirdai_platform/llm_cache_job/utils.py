import os

from pydantic import BaseModel


class InsertLog(BaseModel):
    query: str
    llm_res: str
    reference_hash: str
    num_references: int


class UpdateLogger:
    def __init__(self, log_dir):
        os.makedirs(log_dir, exist_ok=True)
        # We use nomad alloc_id here so that each autoscaling allocation has a distinct file.
        log_file = os.path.join(log_dir, f"{os.getenv('NOMAD_ALLOC_ID')}.jsonl")
        self.stream = open(log_file, "a")

    def log(self, update: BaseModel):
        self.stream.write(update.model_dump_json() + "\n")
        self.stream.flush()
