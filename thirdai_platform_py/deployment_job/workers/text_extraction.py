import json
import logging
import os
import time
from pathlib import Path
from typing import List, Optional
from urllib.parse import urljoin

import requests
import thirdai
import torch
from deployment_job.models.classification_models import TokenClassificationModel
from fastapi.encoders import jsonable_encoder
from platform_common.file_handler import expand_cloud_buckets_and_directories
from platform_common.logging import setup_logger
from platform_common.ndb.ndbv1_parser import parse_doc
from platform_common.pii.data_types.pydantic_models import (
    UnstructuredTokenClassificationResults,
    XMLTokenClassificationResults,
)
from platform_common.pydantic_models.deployment import DeploymentConfig
from platform_common.pydantic_models.training import FileInfo


def load_config():
    with open(os.getenv("CONFIG_PATH")) as file:
        return DeploymentConfig.model_validate_json(file.read())


class UDTRReportProcessorWorker:
    def __init__(self, config: DeploymentConfig):
        self.config = config
        self.reports_base_path = (
            Path(self.config.model_bazaar_dir) / "models" / config.model_id / "reports"
        )
        self.logger = logging.getLogger("UDTRReportProcessorWorker")
        self.logger.setLevel(logging.INFO)

        self.job_endpoint = "http://" + os.getenv("JOB_ENDPOINT") + "/"
        self.auth_header = {"Authorization": f"Bearer {os.environ['JOB_TOKEN']}"}

        self.model = TokenClassificationModel(config=self.config, logger=self.logger)

        cores = int(os.getenv("WORKER_CORES", 1))
        torch.set_num_threads(cores)

    def get_next_report(self):
        try:
            res = requests.post(
                urljoin(self.job_endpoint, "/report/next"), headers=self.auth_header
            )
            if res.status_code == 200:
                data = res.json()["data"]
                report_id = data.get("report_id")
                attempt = data.get("attempt")
                if report_id:
                    self.logger.info(
                        f"Got next report: report_id={report_id} attempt={attempt}"
                    )
                    return report_id, attempt
            self.logger.info("No pending reports available.")
        except Exception as e:
            self.logger.error(f"Error fetching next report: {e}")
        return None, None

    def update_report_status(
        self, report_id: str, new_status: str, attempt: int, msg: Optional[str] = None
    ):
        try:
            res = requests.post(
                urljoin(self.job_endpoint, f"/report/{report_id}/status"),
                headers=self.auth_header,
                json={"new_status": new_status, "attempt": attempt, "msg": msg},
            )
            if res.status_code == 200:
                self.logger.info(
                    f"Updated status for report {report_id} to {new_status}"
                )
            else:
                self.logger.error(
                    f"Error updating status for report {report_id}: {res.content}"
                )
        except Exception as e:
            self.logger.error(f"Exception when updating report {report_id} status: {e}")

    def get_documents(self, report_id: str):
        documents_file = self.reports_base_path / report_id / "documents.json"
        if not documents_file.exists():
            self.logger.error(f"Documents file missing for report {report_id}.")
            raise FileNotFoundError(f"Documents file missing for report {report_id}.")
        with open(documents_file, "r") as file:
            data = json.load(file)

        if isinstance(data, dict):
            documents_json = data.get("documents", [])
            custom_tags = data.get("custom_tags")
        else:
            documents_json = data
            custom_tags = None

        documents = [FileInfo.model_validate(doc) for doc in documents_json]
        documents = expand_cloud_buckets_and_directories(documents)

        if not documents:
            self.logger.error(f"No documents found for report {report_id}.")
            raise ValueError(f"No documents found for report {report_id}.")

        self.logger.info(
            f"Retrieved {len(documents)} document entries for report {report_id}."
        )
        return documents, custom_tags

    def process_documents(
        self,
        documents,
        report_id: str,
        custom_tags: Optional[List[str]] = None,
    ):
        results = []
        for doc in documents:
            self.logger.info(f"Processing document: {doc.path}")
            ndb_doc = parse_doc(
                doc=doc,
                tmp_dir=str(self.reports_base_path / report_id / "documents/tmp"),
            )
            if ndb_doc is None:
                self.logger.error(f"Unable to parse document {doc.path}")
                raise ValueError(
                    f"Unable to process document '{os.path.basename(doc.path)}'. Please ensure that document is a supported type (pdf, docx, csv, html) and has correct extension."
                )

            display_list = ndb_doc.table.df["display"].tolist()
            doc_results = self.model.predict_batch(
                texts=display_list, data_type="unstructured"
            )

            if custom_tags:
                self.logger.info(
                    f"Filtering predictions with custom tags: {custom_tags}"
                )
                filtered_doc_results = []
                for pred in doc_results:

                    if isinstance(pred, UnstructuredTokenClassificationResults):
                        new_tags = [
                            tag if tag in custom_tags else "O"
                            for tag in pred.predicted_tags
                        ]
                        filtered_pred = UnstructuredTokenClassificationResults(
                            data_type=pred.data_type,
                            query_text=pred.query_text,
                            tokens=pred.tokens,
                            predicted_tags=new_tags,
                        )
                        filtered_doc_results.append(filtered_pred)

                    elif isinstance(pred, XMLTokenClassificationResults):
                        new_predictions = [
                            p for p in pred.predictions if p.label in custom_tags
                        ]
                        filtered_pred = XMLTokenClassificationResults(
                            data_type=pred.data_type,
                            query_text=pred.query_text,
                            predictions=new_predictions,
                        )
                        filtered_doc_results.append(filtered_pred)
                    else:
                        filtered_doc_results.append(pred)
                results.append({doc.path: filtered_doc_results})
            else:
                results.append({doc.path: doc_results})
        return results

    def process_report(self, report_id: str, attempt: int):
        try:
            self.logger.info(f"Processing report: {report_id} (attempt {attempt})")
            documents, custom_tags = self.get_documents(report_id)
            if custom_tags:
                self.logger.info(f"Custom tags for report {report_id}: {custom_tags}")

            report_results = self.process_documents(
                documents, report_id, custom_tags=custom_tags
            )

            report_file_path = (
                self.reports_base_path / report_id / f"report_{attempt}.json"
            )
            report_file_path.parent.mkdir(parents=True, exist_ok=True)
            with open(report_file_path, "w") as writer:
                json.dump(
                    {
                        "report_id": report_id,
                        "results": jsonable_encoder(report_results),
                    },
                    writer,
                )

            self.update_report_status(
                report_id=report_id, new_status="complete", attempt=attempt
            )
            self.logger.info(f"Successfully processed report: {report_id}")
        except Exception as e:
            import traceback

            traceback.print_exc()
            self.logger.error(f"Error processing report {report_id}: {e}")
            self.update_report_status(
                report_id=report_id,
                new_status="failed",
                attempt=attempt,
                msg=f"Error: {e}",
            )

    def run(self, poll_interval: int = 5):
        self.logger.info("Starting UDTRReportProcessorWorker...")
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
    print(f"ThirdAI version: {thirdai.__version__}", flush=True)
    config: DeploymentConfig = load_config()

    setup_logger(
        Path(config.model_bazaar_dir) / "logs", f"udt_processor_{config.model_id}"
    )

    worker = UDTRReportProcessorWorker(config=config)
    worker.run()
