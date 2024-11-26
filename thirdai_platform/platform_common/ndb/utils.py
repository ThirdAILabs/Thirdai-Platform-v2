import os
from typing import List

from thirdai import neural_db_v2 as ndb


def delete_docs_and_remove_files(
    db: ndb.NeuralDB, doc_ids: List[str], keep_latest_version: bool = False
):
    deleted_filenames = set([])
    for doc_id in doc_ids:
        deleted_chunks = db.delete_doc(
            doc_id, keep_latest_version=keep_latest_version, return_deleted_chunks=True
        )
        deleted_filenames.update([chunk.document for chunk in deleted_chunks])

    for deleted_filename in deleted_filenames:
        os.remove(deleted_filename)
