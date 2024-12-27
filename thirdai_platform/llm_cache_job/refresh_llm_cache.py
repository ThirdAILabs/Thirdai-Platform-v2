import logging
import os
from pathlib import Path
from typing import List

from cache import NDBSemanticCache
from licensing.verify import verify_license
from llm_cache_job.cache import Cache, NDBSemanticCache
from platform_common.logging import setup_logger
from utils import InsertLog

license_key = os.getenv("LICENSE_KEY")
verify_license.activate_thirdai_license(license_key)

model_bazaar_dir = os.getenv("MODEL_BAZAAR_DIR")
model_id = os.getenv("MODEL_ID")
model_dir = Path(model_bazaar_dir) / "models" / model_id
log_dir: Path = Path(model_bazaar_dir) / "logs" / model_id

insertions_folder = os.path.join(model_dir, "llm_cache", "insertions")


def list_insertions() -> List[InsertLog]:
    insertions = []
    for logfile in os.listdir(insertions_folder):
        if os.path.isfile(logfile) and logfile.endswith(".jsonl"):
            with open(os.path.join(logfile)) as f:
                for line in f.readlines():
                    insertions.extend(InsertLog.model_validate_json(line))
    return insertions


def main():
    setup_logger(log_dir=log_dir, log_prefix="llm-cache")

    logger = logging.getLogger("llm-cache")

    cache: Cache = NDBSemanticCache(model_dir=model_dir, logger=logger)

    insertions = list_insertions()

    for insert_log in insertions:
        cache.insert(insert_log)

    for logfile in os.listdir(insertions_folder):
        os.remove(logfile)


if __name__ == "__main__":
    main()
