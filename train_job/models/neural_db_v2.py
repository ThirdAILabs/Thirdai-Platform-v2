import os
import shutil
import time
import threading
import uuid
from concurrent.futures import ProcessPoolExecutor, as_completed
import multiprocessing as mp
from typing import Any, Dict, List, Optional, Tuple

import thirdai
from config import FileInfo, NDBv2Options, TrainConfig
from fastapi import Response
from models.model import Model
from reporter import Reporter
from thirdai import neural_db_v2 as ndbv2
from utils import (
    check_csv_only,
    check_disk,
    create_s3_client,
    expand_s3_buckets_and_directories,
    get_directory_size,
)


def convert_to_ndb_doc(
    resource_path: str,
    display_path: str,
    doc_id: Optional[str],
    metadata: Optional[Dict[str, Any]],
    options: Dict[str, Any],
) -> ndbv2.Document:
    filename, ext = os.path.splitext(resource_path)

    if ext == ".pdf":
        return ndbv2.PDF(
            resource_path,
            doc_metadata=metadata,
            display_path=display_path,
            doc_id=doc_id,
        )
    elif ext == ".docx":
        return ndbv2.DOCX(
            resource_path,
            doc_metadata=metadata,
            display_path=display_path,
            doc_id=doc_id,
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
            doc_metadata=metadata,
            doc_id=doc_id,
        )
    elif ext == ".csv":
        return ndbv2.CSV(
            resource_path,
            keyword_columns=options.get("csv_strong_columns", []),
            text_columns=options.get("csv_weak_columns", []),
            doc_metadata=metadata,
            display_path=display_path,
            doc_id=doc_id,
        )
    else:
        raise TypeError(f"{ext} Document type isn't supported yet.")


def preload_chunks(
    resource_path: str,
    display_path: str,
    doc_id: Optional[str],
    metadata: Optional[Dict[str, Any]],
    options: Dict[str, Any],
) -> Tuple[ndbv2.Document, str]:
    doc = convert_to_ndb_doc(
        resource_path=resource_path,
        display_path=display_path,
        doc_id=doc_id,
        metadata=metadata,
        options=options,
    )
    return (
        ndbv2.documents.PrebatchedDoc(list(doc.chunks()), doc_id=doc.doc_id()),
        display_path,
    )


def process_file(
    doc: FileInfo, doc_save_dir: str, tmp_dir: str
) -> Tuple[ndbv2.Document, str]:
    """
    Process a file, downloading it from S3 if necessary, and convert it to an NDB file.
    """
    if doc.path.startswith("s3://"):
        print(f"{os.getpid()} Starting to download {doc.path} from s3", flush=True)
        try:
            s3_client = create_s3_client()
            bucket_name, prefix = doc.path.replace("s3://", "").split("/", 1)
            local_file_path = os.path.join(tmp_dir, os.path.basename(prefix))

            s3_client.download_file(bucket_name, prefix, local_file_path)
        except Exception as error:
            print(f"Error downloading file '{doc.path}' from s3 : {error}")
            raise ValueError(f"Error downloading file '{doc.path}' from s3 : {error}")

        print(f"{os.getpid()} Finished downloading {doc.path} from s3", flush=True)
        ndb_doc = preload_chunks(
            resource_path=local_file_path,
            display_path=f"/{bucket_name}.s3.amazonaws.com/{prefix}",
            doc_id=doc.doc_id,
            metadata=doc.metadata,
            options=doc.options,
        )

        print(f"{os.getpid()} Finished loading chunks {doc.path}", flush=True)
        os.remove(local_file_path)
        print(f"{os.getpid()} Returning results {doc.path}", flush=True)
        return ndb_doc

    save_artifact_uuid = str(uuid.uuid4())
    doc_dir = os.path.join(doc_save_dir, save_artifact_uuid)
    os.makedirs(doc_dir, exist_ok=True)
    shutil.copy(src=doc.path, dst=doc_dir)

    return preload_chunks(
        resource_path=os.path.join(doc_dir, os.path.basename(doc.path)),
        display_path=os.path.join(save_artifact_uuid, os.path.basename(doc.path)),
        doc_id=doc.doc_id,
        metadata=doc.metadata,
        options=doc.options,
    )


def parse_docs(docs: List[FileInfo], doc_save_dir: str, tmp_dir: str, task_queue: mp.Queue, job_id: int):
    s3_client = None

    print(f"Starting worker process {job_id}")
    start = time.perf_counter()
    for doc in docs:
        if doc.path.startswith("s3://"):
            try:
                if s3_client is None:
                    s3_client = create_s3_client()
                bucket_name, prefix = doc.path.replace("s3://", "").split("/", 1)
                local_file_path = os.path.join(tmp_dir, os.path.basename(prefix))

                s3_client.download_file(bucket_name, prefix, local_file_path)
            except Exception as error:
                print(f"Error downloading file '{doc.path}' from s3 : {error}")
                raise ValueError(f"Error downloading file '{doc.path}' from s3 : {error}")

            ndb_doc, _ = preload_chunks(
                resource_path=local_file_path,
                display_path=f"/{bucket_name}.s3.amazonaws.com/{prefix}",
                doc_id=doc.doc_id,
                metadata=doc.metadata,
                options=doc.options,
            )
            os.remove(local_file_path)
            task_queue.put(ndb_doc)
        else:
            save_artifact_uuid = str(uuid.uuid4())
            doc_dir = os.path.join(doc_save_dir, save_artifact_uuid)
            os.makedirs(doc_dir, exist_ok=True)
            shutil.copy(src=doc.path, dst=doc_dir)

            ndb_doc, _ = preload_chunks(
                resource_path=local_file_path,
                display_path=f"/{bucket_name}.s3.amazonaws.com/{prefix}",
                doc_id=doc.doc_id,
                metadata=doc.metadata,
                options=doc.options,
            )
            task_queue.put(ndb_doc)
    end = time.perf_counter()

    print(f"Finishing worker process {job_id} parsed {len(docs)} docs in {end-start:.3f} s")


class NeuralDBV2(Model):
    def __init__(self, config: TrainConfig, reporter: Reporter):
        super().__init__(config=config, reporter=reporter)

        self.logger.info(f"THIRDAI VERSION {thirdai.__version__}")

        self.ndb_options: NDBv2Options = self.config.model_options.ndb_options

        if self.config.base_model_id:
            base_model_path = os.path.join(
                self.config.model_bazaar_dir,
                "models",
                self.config.base_model_id,
                "model.ndb",
            )
            self.logger.info(f"Starting training from base model: {base_model_path}")
            shutil.copytree(
                base_model_path,
                self.ndb_save_path(),
                ignore=shutil.ignore_patterns("*.tmpdb"),
                dirs_exist_ok=True,
            )
            self.db = ndbv2.NeuralDB.load(self.ndb_save_path())
        else:
            self.logger.info("Creating new NDBv2 model")
            if self.ndb_options.on_disk:
                save_path = self.ndb_save_path()
            else:
                save_path = None
            self.db = ndbv2.NeuralDB(save_path=save_path)

    def ndb_save_path(self):
        return os.path.join(self.model_dir, "model.ndb")

    def doc_save_path(self):
        return os.path.join(self.ndb_save_path(), "documents")

    def unsupervised_files(self) -> List[FileInfo]:
        return expand_s3_buckets_and_directories(self.config.data.unsupervised_files)

    def supervised_files(self) -> List[FileInfo]:
        all_files = expand_s3_buckets_and_directories(self.config.data.supervised_files)
        check_csv_only(all_files)
        return all_files
    
    def start_parsing_jobs(self, files: List[FileInfo], task_queue: mp.Queue):
        n_jobs = max(1, min(os.cpu_count() - 6, 20))
        chunksize = (len(files) + n_jobs - 1) // n_jobs

        doc_save_dir = self.doc_save_path()
        tmp_dir = self.data_dir / "unsupervised"

        self.logger.info(f"Starting {n_jobs} parsing jobs")

        processes = []
        for i in range(0, len(files), chunksize):
            p = mp.Process(
                target=parse_docs, args=(files[i:i+chunksize], doc_save_dir, tmp_dir, task_queue, len(processes))
            )
            p.start()
            processes.append(p)

        for p in processes:
            p.join()

        task_queue.put(None)
        self.logger.info("Parsing jobs completed")

    def unsupervised_train(self, files: List[FileInfo], batch_size=100):
        self.logger.info("Starting unsupervised training.")

        os.makedirs(self.data_dir / "unsupervised", exist_ok=True)

        docs_indexed = 0
        
        task_queue = mp.Queue(maxsize=4 * batch_size)
        
        parsing_jobs = threading.Thread(target=self.start_parsing_jobs, args=(files, task_queue))
        parsing_jobs.start()

        batch = []
        while 1:
            doc = task_queue.get()
            if doc is None:
                if batch:
                    self.db.insert(batch)
                    docs_indexed += len(batch)
                    self.logger.info(f"Inserted {docs_indexed} docs")
                break
            batch.append(doc)
            if len(batch) == batch_size:
                self.db.insert(batch)
                docs_indexed += len(batch)
                batch.clear()
                self.logger.info(f"Inserted {docs_indexed} docs")

        parsing_jobs.join()

        self.logger.info("Completed unsupervised training.")


    # def unsupervised_train(self, files: List[FileInfo], batch_size=100):
    #     self.logger.info("Starting unsupervised training.")

    #     os.makedirs(self.data_dir / "unsupervised", exist_ok=True)

    #     num_workers = max(1, min(os.cpu_count() - 6, 20))
    #     self.logger.info(f"Starting {num_workers} parsing processes")

    #     doc_save_dir = self.doc_save_path()
    #     tmp_dir = self.data_dir / "unsupervised"

    #     docs_indexed = 0

    #     with ProcessPoolExecutor(max_workers=num_workers) as executor:
    #         futures = [
    #             executor.submit(process_file, file, doc_save_dir, tmp_dir)
    #             for file in files
    #         ]

    #         self.logger.info(f"n_futures = {len(futures)}")

    #         batch = []
    #         for future in as_completed(futures):
    #             try:
    #                 ndb_doc, filename = future.result()
    #                 batch.append(ndb_doc)
    #                 self.logger.debug(
    #                     f"Parsed document: doc_id={ndb_doc.doc_id()} {filename=}"
    #                 )
    #             except Exception as e:
    #                 self.logger.error(f"Error processing file: {e}")

    #             if len(batch) == batch_size:
    #                 self.db.insert(batch)
    #                 docs_indexed += len(batch)
    #                 self.logger.info(f"Inserted {docs_indexed} docs")
    #                 batch.clear()

    #         if len(batch) > 0:
    #             self.db.insert(batch)
    #             docs_indexed += len(batch)
    #             self.logger.info(f"Inserted {docs_indexed} docs")

    #     self.logger.info("Completed unsupervised training.")

    def supervised_train(self, files: List[FileInfo]):
        self.logger.info("Starting supervised training.")

        for file in files:
            supervised_dataset = ndbv2.supervised.CsvSupervised(
                file.path,
                query_column=file.options.get("csv_query_column"),
                id_column=file.options.get("csv_id_column"),
                id_delimiter=file.options.get("csv_id_delimiter"),
            )

            self.db.supervised_train(supervised_dataset)

        self.logger.info("Completed supervised training.")

    def train(self, **kwargs):
        """
        Train the NeuralDB with unsupervised and supervised data.
        """
        self.logger.info("Training process started.")
        self.reporter.report_status(self.config.model_id, "in_progress")

        s = time.perf_counter()
        unsupervised_files = self.unsupervised_files()
        e = time.perf_counter()
        self.logger.info(
            f"Listed {len(unsupervised_files)} unsupervised files in {e-s:.4f} seconds"
        )

        s = time.perf_counter()
        supervised_files = self.supervised_files()
        e = time.perf_counter()
        self.logger.info(
            f"Listed {len(supervised_files)} supervised files in {e-s:.4f} seconds"
        )

        start_time = time.time()

        if unsupervised_files:
            self.logger.info(f"Found {len(unsupervised_files)} unsupervised files.")
            check_disk(self.db, self.config.model_bazaar_dir, unsupervised_files)
            self.unsupervised_train(unsupervised_files)
            self.logger.info("Completed Unsupervised Training")

        if supervised_files:
            self.logger.info(f"Found {len(supervised_files)} supervised files.")
            check_disk(self.db, self.config.model_bazaar_dir, supervised_files)
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
        if not self.ndb_options.on_disk:
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
            model_id=self.config.model_id,
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
