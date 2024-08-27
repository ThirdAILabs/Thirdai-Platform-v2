import ast
import json
import os
import queue
import shutil
import threading
import time
import uuid
from concurrent.futures import ProcessPoolExecutor, as_completed
from typing import List, Tuple

import thirdai
from fastapi import Response
from model import Model
from thirdai import neural_db_v2 as ndbv2
from utils import check_disk, create_s3_client, get_directory_size, list_files
from variables import FinetunableRetrieverVariables


def convert_to_ndb_doc(resource_path: str, display_path: str) -> ndbv2.Document:
    filename, ext = os.path.splitext(resource_path)

    json_file_path = f"{filename}_metadata.json"
    doc_metadata = None

    if os.path.exists(json_file_path):
        with open(json_file_path, "r") as json_file:
            doc_metadata = json.load(json_file)
    if ext == ".pdf":
        return ndbv2.PDF(
            resource_path, doc_metadata=doc_metadata, display_path=display_path
        )
    elif ext == ".docx":
        return ndbv2.DOCX(
            resource_path, doc_metadata=doc_metadata, display_path=display_path
        )
    elif ext == ".html":
        with open(resource_path, "r", encoding="utf-8") as f:
            html_content = f.read()

        dummy_response = Response()
        dummy_response.status_code = 200
        dummy_response._content = html_content.encode("utf-8")

        return ndbv2.URL(
            os.path.basename(filename),
            response=dummy_response,
            doc_metadata=doc_metadata,
        )
    elif ext == ".csv":
        return ndbv2.CSV(
            resource_path,
            id_column=os.getenv("CSV_ID_COLUMN", None),
            keyword_columns=ast.literal_eval(os.getenv("CSV_STRONG_COLUMNS", "None")),
            text_columns=ast.literal_eval(os.getenv("CSV_WEAK_COLUMNS", "None")),
            doc_metadata=doc_metadata,
            display_path=display_path,
        )
    else:
        raise TypeError(f"{ext} Document type isn't supported yet.")


def preload_chunks(resource_path: str, display_path: str) -> Tuple[ndbv2.Document, str]:
    # TODO(V2 Support): Add an option for users to set the doc_id
    doc = convert_to_ndb_doc(resource_path=resource_path, display_path=display_path)
    return ndbv2.documents.PrebatchedDoc(doc.chunks(), doc_id=doc.doc_id), resource_path


def process_file(
    file: str, doc_save_dir: str, tmp_dir: str
) -> Tuple[ndbv2.Document, str]:
    """
    Process a file, downloading it from S3 if necessary, and convert it to an NDB file.
    """
    if file.startswith("s3://"):
        s3_client = create_s3_client()
        bucket_name, prefix = file.replace("s3://", "").split("/", 1)
        local_file_path = os.path.join(tmp_dir, os.path.basename(prefix))

        # Download the file from S3
        try:
            s3_client.download_file(bucket_name, prefix, local_file_path)
        except Exception as error:
            print(f"There was an error downloading the file from s3 : {error}. {file}")
            return f"There was an error downloading the file from s3 : {error}"

        doc = preload_chunks(
            resource_path=local_file_path,
            display_path=f"/{bucket_name}.s3.amazonaws.com/{prefix}",
        )
        os.remove(local_file_path)

        return doc

    save_artifact_uuid = uuid.uuid4()
    doc_dir = os.path.join(doc_save_dir, save_artifact_uuid)
    os.makedirs(doc_dir, exist_ok=True)
    shutil.copy(src=file, dst=doc_dir)

    return preload_chunks(
        resource_path=os.path.join(doc_dir, os.path.basename(file)),
        display_path=os.path.join(save_artifact_uuid, os.path.basename(file)),
    )


class NeuralDBV2(Model):
    def __init__(self):
        super().__init__()

        self.retriever_variables: FinetunableRetrieverVariables = (
            FinetunableRetrieverVariables.load_from_env()
        )

        if self.retriever_variables.on_disk:
            save_path = self.ndb_save_path()
        else:
            save_path = None
        self.db = ndbv2.NeuralDB(save_path=save_path)

    def ndb_save_path(self):
        return os.path.join(self.model_dir, "model.ndb")

    def doc_save_path(self):
        return os.path.join(self.ndb_save_path(), "documents")

    def parser(
        self, files: list[str], task_queue: queue.Queue, doc_save_dir: str, tmp_dir: str
    ):
        max_cores = os.cpu_count()
        num_workers = max(1, max_cores - 6)
        with ProcessPoolExecutor(max_workers=num_workers) as executor:
            futures = [
                executor.submit(process_file, file, doc_save_dir, tmp_dir)
                for file in files
            ]

            for future in as_completed(futures):
                try:
                    ndb_file, filename = future.result()
                    if ndb_file and not isinstance(ndb_file, str):
                        task_queue.put(ndb_file)
                        print(f"Successfully processed {filename}", flush=True)
                except Exception as e:
                    print(f"Error processing file {filename}: {e}")

    def indexer(self, task_queue: queue.Queue, batch_size: int):
        batch = []

        while True:
            ndb_doc = task_queue.get()
            if ndb_doc is None:
                if batch:
                    self.db.insert(batch)
                break

            batch.append(ndb_doc)

            if len(batch) >= batch_size and task_queue.qsize() == 0:
                self.db.insert(batch)
                batch.clear()

    def unsupervised_train(self, files: List[str]):
        self.logger.info("Starting unsupervised training.")
        task_queue = queue.Queue()

        parser_thread = threading.Thread(
            target=self.parser,
            args=(
                files,  # files
                task_queue,  # task_queue
                self.doc_save_path(),  # doc_save_dir
                self.data_dir / "unsupervised",  # tmp_dir
            ),
        )

        indexer_thread = threading.Thread(
            target=self.indexer,
            args=(task_queue, 50),
        )

        self.logger.info(
            "Starting parser and indexer threads for unsupervised training."
        )
        parser_thread.start()
        indexer_thread.start()

        parser_thread.join()
        task_queue.put(None)  # Signal the indexer to exit
        parser_thread.join()
        self.logger.info("Completed unsupervised training.")

    def supervised_train(self, files: List[str]):
        self.logger.info("Starting supervised training.")

        for file in files:
            supervised_dataset = ndbv2.supervised.CsvSupervised(
                file,
                query_column=os.getenv("CSV_QUERY_COLUMN", None),
                id_column=os.getenv("CSV_ID_COLUMN", None),
                id_delimiter=os.getenv("CSV_ID_DELIMITER", None),
            )

            self.db.supervised_train(supervised_dataset)

        self.logger.info("Completed supervised training.")

    def train(self, **kwargs):
        """
        Train the NeuralDB with unsupervised and supervised data.
        """
        self.logger.info("Training process started.")
        self.reporter.report_status(self.general_variables.model_id, "in_progress")

        unsupervised_files = list_files(self.data_dir / "unsupervised")
        supervised_files = list_files(self.data_dir / "supervised")

        start_time = time.time()

        if unsupervised_files:
            self.logger.info(f"Found {len(unsupervised_files)} unsupervised files.")
            check_disk(
                self.db, self.general_variables.model_bazaar_dir, unsupervised_files
            )
            self.unsupervised_train(unsupervised_files)
            self.logger.info("Completed Unsupervised Training")

        if supervised_files:
            self.logger.info(f"Found {len(supervised_files)} supervised files.")
            check_disk(
                self.db, self.general_variables.model_bazaar_dir, supervised_files
            )
            self.supervised_train(supervised_files)
            self.logger.info("Completed Supervised Training")

        train_time = time.time() - start_time
        self.logger.info(f"Total training time: {train_time} seconds")

        self.save()
        self.logger.info("Model saved successfully.")

        self.finalize_training(train_time)
        self.logger.info("Training finalized successfully.")

    def evaluate(self, **kwargs):
        """
        Evaluate the NeuralDB. Not implemented.
        """
        self.logger.warning("Evaluation method called. Not implemented.")

    def save(self):
        if not self.retriever_variables.on_disk:
            self.db.save(self.ndb_save_path())

    def get_latency(self) -> float:
        self.logger.info("Measuring latency of the NeuralDBv2 instance.")
        start_time = time.time()

        self.db.search("Checking for latency", top_k=5)

        latency = time.time() - start_time
        self.logger.info(f"Latency measured: {latency} seconds.")
        return latency

    def get_size_in_memory(self) -> int:
        def get_size(path):
            if os.path.isfile(path):
                return os.stat(path).st_size
            return get_directory_size(path)

        # TODO(Nicholas): update this calculation for on_disk=True
        size_in_memory = int(
            get_size(self.db.retriever_path(self.ndb_save_path())) * 1.5
            + get_size(self.db.chunk_store_path(self.ndb_save_path()))
        )
        self.logger.info(f"Size of the model in memory: {size_in_memory} bytes")
        return size_in_memory

    def finalize_training(self, train_time: int):
        self.logger.info("Finalizing training process.")

        self.reporter.report_complete(
            model_id=self.general_variables.model_id,
            metadata={
                "num_params": str(self.db.retriever.retriever.size()),
                "size": str(get_directory_size(self.ndb_save_path())),
                "size_in_memory": str(self.get_size_in_memory()),
                "thirdai_version": str(thirdai.__version__),
                "training_time": str(train_time),
                "latency": str(self.get_latency()),
            },
        )
        self.logger.info("Training finalized and reported successfully.")
