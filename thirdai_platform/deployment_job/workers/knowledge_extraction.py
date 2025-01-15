import json
import logging
import os
import time
from pathlib import Path
from typing import Optional
from urllib.parse import urljoin

import requests
import thirdai
import torch
from licensing.verify import verify_license
from platform_common.file_handler import expand_cloud_buckets_and_directories
from platform_common.logging import setup_logger
from platform_common.ndb.ndbv2_parser import parse_doc
from platform_common.pydantic_models.deployment import DeploymentConfig
from platform_common.pydantic_models.training import FileInfo
from thirdai import neural_db_v2 as ndb


def load_config():
    with open(os.getenv("CONFIG_PATH")) as file:
        return DeploymentConfig.model_validate_json(file.read())


KE_PROMPT = (
    "Given this context, act as a financial expert and give a short answer "
    "for the following question based on the provided context in an unbiased, "
    "comprehensive and scholarly tone:"
)


class ReportProcessorWorker:
    def __init__(self, config: DeploymentConfig):
        self.config = config
        self.reports_base_path = (
            Path(self.config.model_bazaar_dir) / "models" / config.model_id / "reports"
        )
        self.logger = logging.getLogger("ReportProcessorWorker")

        self.llm_endpoint = urljoin(
            self.config.model_bazaar_endpoint, "llm-dispatch/generate"
        )

        cores = int(os.getenv("WORKER_CORES"))
        if hasattr(thirdai, "set_global_num_threads"):
            thirdai.set_global_num_threads(cores)
        torch.set_num_threads(cores)

        self.job_endpoint = "http://" + os.getenv("JOB_ENDPOINT") + "/"
        self.logger.info(f"JOB ENDPOINT: '{self.job_endpoint}'")

        self.auth_header = {"Authorization": f"Bearer {os.environ['JOB_TOKEN']}"}

        self.llm_provider = self.config.options["llm_provider"]
        self.genai_key = self.config.options["genai_key"]
        self.advanced_indexing = self.config.options["advanced_indexing"] == "true"
        self.rerank = self.config.options["rerank"] == "true"
        self.generate_answers = self.config.options["generate_answers"] == "true"

        self.logger.info(
            f"options: advanced_indexing={self.advanced_indexing} rerank={self.rerank} generate_answers={self.generate_answers}"
        )

        verify_license.activate_thirdai_license(self.config.license_key)

    def get_next_report(self):
        res = requests.post(
            urljoin(self.job_endpoint, "/report/next"), headers=self.auth_header
        )

        if res.status_code == 200:
            data = res.json()["data"]
            self.logger.info(
                f"got next report from queue: report_id={data['report_id']} attempt={data['attempt']}"
            )
            return data["report_id"], data["attempt"]

        self.logger.error(f"error retreiving next report: {str(res.content)}")
        return None, None

    def update_report_status(
        self, report_id: str, new_status: str, attempt: int, msg: Optional[str] = None
    ):
        res = requests.post(
            urljoin(self.job_endpoint, f"/report/{report_id}/status"),
            headers=self.auth_header,
            json={"new_status": new_status, "attempt": attempt, "msg": msg},
        )

        if res.status_code == 200:
            self.logger.info(f"updated status of report {report_id} to {new_status}")
        else:
            self.logger.error(f"error updating report status: {str(res.content)}")

    def get_questions(self):
        res = requests.get(
            urljoin(self.job_endpoint, "/questions-internal"), headers=self.auth_header
        )
        if res.status_code == 200:
            self.logger.info("successfully got list of questions")
            return res.json()["data"]

        raise ValueError(f"error getting list of questions: {str(res.content)}")

    def get_documents(self, report_id: str):
        documents_file = self.reports_base_path / report_id / "documents.json"

        if not documents_file.exists():
            self.logger.error(f"Documents file missing for report {report_id}.")
            raise FileNotFoundError(f"Documents file missing for report {report_id}.")

        with open(documents_file) as file:
            documents = json.load(file)

        documents = [FileInfo.model_validate(doc) for doc in documents]
        documents = expand_cloud_buckets_and_directories(documents)

        if not documents:
            self.logger.error(f"No documents found for report {report_id}.")
            raise ValueError(f"No documents found for report {report_id}.")

        return documents

    def create_ndb(self, documents, report_id: str):
        os.makedirs(self.reports_base_path / report_id / "documents/tmp", exist_ok=True)

        self.logger.info("starting document parsing")
        s = time.perf_counter()
        docs = []
        for doc in documents:
            self.logger.info(f"parsing document: {doc.path}")
            ndb_doc = parse_doc(
                doc=doc,
                doc_save_dir=str(self.reports_base_path / report_id / "documents"),
                tmp_dir=str(self.reports_base_path / report_id / "documents/tmp"),
            )
            if ndb_doc is None:
                self.logger.error(f"unable to parse document {doc.path}")
                raise ValueError(
                    f"Unable to process document '{os.path.basename(doc.path)}'. Please ensure that document is a supported type (pdf, docx, csv, html) and has correct extension."
                )
            else:
                docs.append(ndb_doc)
                self.logger.info(f"parsed document: {doc.path}")

        total_chunks = 0
        for doc in docs:
            for chunk in doc.chunks():
                total_chunks += len(chunk.text)

        e = time.perf_counter()
        self.logger.info(
            f"document parsing complete: time={e-s:.3f}s total_chunks={total_chunks}"
        )

        db = ndb.NeuralDB(splade=(total_chunks < 5000) and self.advanced_indexing)

        self.logger.info("starting indexing")
        s = time.perf_counter()
        db.insert(docs)
        e = time.perf_counter()
        self.logger.info(
            f"indexing complete: time={e-s:.3f}s ndocs={len(docs)} chunks={db.retriever.retriever.size()}"
        )

        return db

    def answer_questions(self, db, questions):
        queries = []
        for question in questions:
            query = question["question_text"] + " " + " ".join(question["keywords"])
            queries.append(query)

        s = time.perf_counter()
        self.logger.info("starting answer generation")
        search_results = db.search_batch(queries, top_k=5, rerank=self.rerank)

        search_results = [
            [{"text": chunk.text, "source": chunk.document} for chunk, _ in refs]
            for refs in search_results
        ]

        if self.generate_answers:
            answers = [
                self.generate(q["question_text"], refs)
                for q, refs in zip(questions, search_results)
            ]
        else:
            answers = [None] * len(search_results)

        report_results = [
            {
                "question_id": question["question_id"],
                "question": question["question_text"],
                "answer": answer,
                "references": refs,
            }
            for question, answer, refs in zip(questions, answers, search_results)
        ]

        e = time.perf_counter()
        self.logger.info(
            f"answer generation complete: time={e-s:.3f}s n_questions={len(questions)}"
        )

        return report_results

    def process_report(self, report_id: str, attempt: int):
        try:
            self.logger.info(f"Processing report: {report_id=} attempt {attempt=}")

            documents = self.get_documents(report_id)

            questions = self.get_questions()

            db = self.create_ndb(documents=documents, report_id=report_id)

            report_results = self.answer_questions(db=db, questions=questions)

            report_file_path = (
                self.reports_base_path / report_id / f"report_{attempt}.json"
            )
            report_file_path.parent.mkdir(parents=True, exist_ok=True)
            with open(report_file_path, mode="w") as writer:
                json.dump({"report_id": report_id, "results": report_results}, writer)

            self.update_report_status(
                report_id=report_id, new_status="complete", attempt=attempt
            )
            self.logger.info(f"Successfully processed report: {report_id}")

        except Exception as e:
            self.logger.error(f"Error processing report {report_id}: {e}")
            self.update_report_status(
                report_id=report_id,
                new_status="failed",
                attempt=attempt,
                msg=f"Error processing report: {e}",
            )

    def generate(self, question, references):
        self.logger.debug(f"generating answer for question: {question}")
        response = requests.post(
            self.llm_endpoint,
            headers={
                "Content-Type": "application/json",
            },
            json={
                "query": question,
                "task_prompt": KE_PROMPT,
                "references": references,
                "key": self.genai_key,
                "provider": self.llm_provider,
            },
        )

        if response.status_code != 200:
            self.logger.error(
                f"Not able to get generated answer for question {question} status_code: {response.status_code}"
            )
            return "error generating answer"

        self.logger.debug(f"generated answer for question: {question}")
        return response.text

    def run(self, poll_interval: int = 5):
        self.logger.info("Starting ReportProcessorWorker...")

        while True:
            try:
                report_id, attempt = self.get_next_report()
                if not report_id:
                    self.logger.info("No pending reports. Sleeping...")
                    time.sleep(poll_interval)
                    continue

                self.process_report(report_id, attempt)

            except Exception as e:
                self.logger.error(f"Worker encountered an error: {e}")
                time.sleep(poll_interval)


if __name__ == "__main__":
    config: DeploymentConfig = load_config()

    setup_logger(
        Path(config.model_bazaar_dir) / "logs",
        f"knowledge_extraction_{config.model_id}",
    )

    worker = ReportProcessorWorker(config=config)
    worker.run()
