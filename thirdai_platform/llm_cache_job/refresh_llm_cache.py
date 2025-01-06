import logging
import os
import shutil
from pathlib import Path
from typing import List

from licensing.verify import verify_license
from llm_cache_job.cache import Cache, NDBSemanticCache
from llm_cache_job.reporter import HttpReporter
from llm_cache_job.utils import InsertLog
from platform_common.logging import setup_logger

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

    reporter = HttpReporter(os.getenv("MODEL_BAZAAR_ENDPOINT"), logger)

    reporter.report_status(model_id, "in_progress")

    try:
        # TODO does this work while deploying?
        original_cache_ndb_path = os.path.join(model_dir, "llm_cache", "llm_cache.ndb")
        new_cache_ndb_path = os.path.join(model_dir, "llm_cache", "new_llm_cache.ndb")
        shutil.copytree(
            original_cache_ndb_path,
            new_cache_ndb_path,
            ignore=shutil.ignore_patterns("*.tmpdb"),
            dirs_exist_ok=True,
        )
        cache: Cache = NDBSemanticCache(
            cache_ndb_path=new_cache_ndb_path, log_dir=model_dir, logger=logger
        )

        insertions = list_insertions()

        cache.insert(insertions)

        for logfile in os.listdir(insertions_folder):
            os.remove(os.path.join(insertions_folder, logfile))

        shutil.copytree(
            new_cache_ndb_path,
            original_cache_ndb_path,
            ignore=shutil.ignore_patterns("*.tmpdb"),
            dirs_exist_ok=True,
        )
        os.remove(new_cache_ndb_path)
    except Exception as e:
        reporter.report_status(model_id, "failed", str(e))

    reporter.report_status(model_id, "complete")


if __name__ == "__main__":
    main()
