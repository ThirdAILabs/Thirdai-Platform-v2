import json
import os
import shutil
import time
import uuid
from datetime import datetime, timedelta
from pathlib import Path
from typing import List, Optional

from deployment_job.models.classification_models import (
    ClassificationModel,
    TextClassificationModel,
    TokenClassificationModel,
)
from deployment_job.permissions import Permissions
from deployment_job.pydantic_models.inputs import (
    DocumentList,
    TextAnalysisPredictParams,
    TokenAnalysisPredictParams,
)
from deployment_job.reporter import Reporter
from deployment_job.routers.knowledge_extraction import JOB_TOKEN
from deployment_job.throughput import Throughput
from fastapi import (
    APIRouter,
    Depends,
    Form,
    HTTPException,
    Query,
    Request,
    UploadFile,
    status,
)
from fastapi.encoders import jsonable_encoder
from platform_common.dependencies import is_on_low_disk
from platform_common.file_handler import download_local_files
from platform_common.logging import JobLogger, LogCode
from platform_common.ndb.ndbv1_parser import convert_to_ndb_file
from platform_common.pii.data_types import (
    UnstructuredTokenClassificationResults,
    XMLTokenClassificationResults,
)
from platform_common.pii.schema import Base, UDTReport
from platform_common.pydantic_models.deployment import DeploymentConfig
from platform_common.thirdai_storage.data_types import (
    LabelCollection,
    LabelStatus,
    TextClassificationData,
    TokenClassificationData,
)
from platform_common.utils import response
from prometheus_client import Summary
from pydantic import BaseModel, ValidationError
from sqlalchemy import and_, create_engine, or_
from sqlalchemy.orm import Session, scoped_session, sessionmaker
from thirdai import neural_db as ndb

udt_predict_metric = Summary("udt_predict", "UDT predictions")

udt_query_length = Summary("udt_query_length", "Distribution of query lengths")


def verify_job_token(request: Request):
    auth_header = request.headers.get("Authorization")
    if auth_header.split(" ", 1)[1] != JOB_TOKEN:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="invalid access token",
            headers={"WWW-Authenticate": "Bearer"},
        )
    return None


MAX_ATTEMPTS = 3
REPORT_TIMEOUT = timedelta(minutes=10)


class UDTBaseRouter:
    def __init__(self, config: DeploymentConfig, reporter: Reporter, logger: JobLogger):
        self.model: ClassificationModel = self.get_model(config, logger)
        self.logger = logger

        # TODO(Nicholas): move these metrics to prometheus
        self.start_time = time.time()
        self.tokens_identified = Throughput()
        self.queries_ingested = Throughput()
        self.queries_ingested_bytes = Throughput()

        self.router = APIRouter()

        self.reports_base_path = (
            Path(config.model_bazaar_dir) / "models" / config.model_id / "reports"
        )
        self.reports_base_path.mkdir(parents=True, exist_ok=True)

        self.db_path = (
            Path(config.model_bazaar_dir)
            / "models"
            / config.model_id
            / "reports"
            / "report.db"
        )

        self.engine = self._initialize_db()
        self.Session = scoped_session(sessionmaker(bind=self.engine))

        self.router.add_api_route("/stats", self.stats, methods=["GET"])
        self.router.add_api_route("/get-text", self.get_text, methods=["POST"])

    def _initialize_db(self):
        if not self.db_path.parent.exists():
            self.db_path.parent.mkdir(parents=True)

        engine = create_engine(f"sqlite:///{self.db_path}")
        Base.metadata.create_all(engine)
        return engine

    @staticmethod
    def get_model(config: DeploymentConfig, logger: JobLogger) -> ClassificationModel:
        raise NotImplementedError("Subclasses should implement this method")

    def get_text(
        self,
        file: UploadFile,
        _=Depends(Permissions.verify_permission("read")),
    ):
        """
        Process an uploaded file to extract text content.

        Args:
            file (UploadFile): The uploaded file to process.
            _ (str): Unused parameter for permission check dependency injection.

        Returns:
            A JSON response containing the extracted text content.
        """
        try:
            # Define the path where the uploaded file will be temporarily saved
            destination_path = self.model.data_dir / file.filename

            # Save the uploaded file to the temporary location
            with open(destination_path, "wb") as f:
                f.write(file.file.read())

            # Convert the file to an ndb Document object
            # This likely involves parsing and processing the file content
            doc: ndb.Document = convert_to_ndb_file(
                destination_path, metadata=None, options=None
            )

            # Extract the 'display' column from the document's table
            # and convert it to a list
            display_list = doc.table.df["display"].tolist()

            # Remove the temporary file
            os.remove(destination_path)

            self.logger.info(
                f"Processing text extraction for file: {file.filename}",
                code=LogCode.FILE_VALIDATION,
            )

            # Return a JSON response with the extracted text content
            return response(
                status_code=status.HTTP_200_OK,
                message="Successful",
                data=jsonable_encoder(display_list),
            )
        except Exception as e:
            self.logger.error(
                f"Error processing text extraction for file: {file.filename}. Error: {e}",
                code=LogCode.FILE_VALIDATION,
            )
            raise e

    def stats(self, _=Depends(Permissions.verify_permission("read"))):
        """
        Returns statistics about the deployment such as the number of tokens identified, number of
        queries ingested, and total size of queries ingested.

        Parameters:
        - token: str - Authorization token (inferred from permissions dependency).

        Returns:
        - JSONResponse: Statistics about deployment usage. Example response:
        {
            "past_hour": {
                "tokens_identified": 125,
                "queries_ingested": 12,
                "queries_ingested_bytes": 7223,
            },
            "total": {
                "tokens_identified": 1125,
                "queries_ingested": 102,
                "queries_ingested_bytes": 88101,
            },
            "uptime": 35991
        }
        uptime is given in seconds.
        """
        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data={
                "past_hour": {
                    "tokens_identified": self.tokens_identified.past_hour(),
                    "queries_ingested": self.queries_ingested.past_hour(),
                    "queries_ingested_bytes": self.queries_ingested_bytes.past_hour(),
                },
                "total": {
                    "tokens_identified": self.tokens_identified.past_hour(),
                    "queries_ingested": self.queries_ingested.past_hour(),
                    "queries_ingested_bytes": self.queries_ingested_bytes.past_hour(),
                },
                "uptime": int(time.time() - self.start_time),
            },
        )

    def get_session(self) -> Session:
        return self.Session()


class UDTRouterTextClassification(UDTBaseRouter):
    def __init__(self, config: DeploymentConfig, reporter: Reporter, logger: JobLogger):
        super().__init__(config, reporter, logger)
        # Add routes specific to text classification
        self.router.add_api_route(
            "/insert_sample",
            self.insert_sample,
            methods=["POST"],
            dependencies=[Depends(is_on_low_disk(path=config.model_bazaar_dir))],
        )
        self.router.add_api_route(
            "/get_recent_samples", self.get_recent_samples, methods=["GET"]
        )
        self.router.add_api_route("/predict", self.predict, methods=["POST"])

    @udt_predict_metric.time()
    def predict(
        self,
        params: TextAnalysisPredictParams,
        _=Depends(Permissions.verify_permission("read")),
    ):
        """
        Predicts the output based on the provided query parameters.

        Parameters:
        - text: str - The text for the sample to perform inference on.
        - top_k: int - The number of results to return.
        - token: str - Authorization token (inferred from permissions dependency).

        Returns:
        - JSONResponse: Prediction results.

        Example Request Body:
        ```
        {
            "text": "What is artificial intelligence?",
            "top_k": 5
        }
        ```
        """
        start_time = time.perf_counter()

        text_length = len(params.text.split())
        udt_query_length.observe(text_length)

        results = self.model.predict(**params.model_dump())
        self.queries_ingested.log(1)
        self.queries_ingested_bytes.log(len(params.text))

        end_time = time.perf_counter()
        time_taken = end_time - start_time

        # Add time_taken to the response data
        response_data = {
            "prediction_results": jsonable_encoder(results),
            "time_taken": time_taken,
        }

        self.logger.debug(f"Prediction complete with time taken: {time_taken} seconds")

        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=response_data,
        )

    @staticmethod
    def get_model(config: DeploymentConfig, logger: JobLogger) -> ClassificationModel:
        logger.info(f"Initializing Nlp Text Classification model")
        return TextClassificationModel(config=config, logger=logger)

    def insert_sample(
        self,
        sample: TextClassificationData,
        _=Depends(Permissions.verify_permission("write")),
    ):
        self.model.insert_sample(sample)
        return response(status_code=status.HTTP_200_OK, message="Successful")

    def get_recent_samples(
        self,
        num_samples: int = Query(
            default=5, ge=1, le=100, description="Number of recent samples to retrieve"
        ),
        _=Depends(Permissions.verify_permission("read")),
    ):
        recent_samples = self.model.get_recent_samples(num_samples=num_samples)
        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=jsonable_encoder(recent_samples),
        )


class UpdateReportRequest(BaseModel):
    new_status: str
    attempt: int
    msg: Optional[str] = (None,)


class UDTRouterTokenClassification(UDTBaseRouter):
    def __init__(self, config: DeploymentConfig, reporter: Reporter, logger: JobLogger):
        super().__init__(config, reporter, logger)
        # The following routes are only applicable for token classification models
        # TODO(Shubh): Make different routers for text and token classification models
        self.router.add_api_route("/add_labels", self.add_labels, methods=["POST"])
        self.router.add_api_route(
            "/insert_sample",
            self.insert_sample,
            methods=["POST"],
            dependencies=[Depends(is_on_low_disk(path=config.model_bazaar_dir))],
        )
        self.router.add_api_route("/get_labels", self.get_labels, methods=["GET"])
        self.router.add_api_route(
            "/get_recent_samples", self.get_recent_samples, methods=["GET"]
        )
        self.router.add_api_route("/predict", self.predict, methods=["POST"])
        self.router.add_api_route(
            "/report/create",
            self.new_report,
            methods=["POST"],
            dependencies=[Depends(is_on_low_disk(path=self.config.model_bazaar_dir))],
        )
        self.router.add_api_route(
            "/report/{report_id}", self.get_report, methods=["GET"]
        )
        self.router.add_api_route(
            "/report/{report_id}", self.delete_report, methods=["DELETE"]
        )
        self.router.add_api_route("/reports", self.list_reports, methods=["GET"])
        self.router.add_api_route(
            "/report/{report_id}/status", self.update_report_status, methods=["POST"]
        )
        self.router.add_api_route(
            "/report/next", self.next_unprocessed_report, methods=["POST"]
        )

    @staticmethod
    def get_model(config: DeploymentConfig, logger: JobLogger) -> ClassificationModel:
        logger.info(f"Initializing Nlp Token Classification model")
        return TokenClassificationModel(config=config, logger=logger)

    def add_labels(
        self,
        labels: LabelCollection,
        _=Depends(Permissions.verify_permission("write")),
    ):
        """
        Adds new labels to the model.
        Parameters:
        - labels: LabelEntityList - A list of LabelEntity specifying the name of the label and description for generating synthetic data for the label.
        - token: str - Authorization token (inferred from permissions dependency).
        Returns:
        - JSONResponse: Status specifying whether or not the request to add labels was successful.

        Example Request Body:
        ```
        {
            "tags": [
                {
                    "name": "label1",
                    "description": "Description for label1"
                },
                {
                    "name": "label2",
                    "description": "Description for label2"
                }
            ]
        }
        ```
        """

        for label in labels.tags:
            assert label.status == LabelStatus.uninserted
        self.model.add_labels(labels)
        return response(status_code=status.HTTP_200_OK, message="Successful")

    def insert_sample(
        self,
        sample: TokenClassificationData,
        _=Depends(Permissions.verify_permission("write")),
    ):
        """
        Inserts a sample into the model.
        Parameters:
        - sample: TokenClassificationSample - The sample to insert into the model.
        - token: str - Authorization token (inferred from permissions dependency).
        Returns:
        - JSONResponse: Status specifying whether or not the request to insert a sample was successful.

        Example Request Body:
        ```
        {
            "tokens": ["This", "is", "a", "test", "sample"],
            "tags": ["O", "O", "O", "test_label", "O"]
        }
        ```
        """
        self.model.insert_sample(sample)
        return response(status_code=status.HTTP_200_OK, message="Successful")

    def get_labels(self, _=Depends(Permissions.verify_permission("read"))):
        """
        Retrieves the labels from the model.
        Parameters:
        - token: str - Authorization token (inferred from permissions dependency).
        Returns:
        - JSONResponse: The labels from the model.
        """
        labels = self.model.get_labels()
        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=jsonable_encoder(labels),
        )

    def get_recent_samples(
        self,
        num_samples: int = Query(
            default=5, ge=1, le=100, description="Number of recent samples to retrieve"
        ),
        _=Depends(Permissions.verify_permission("read")),
    ):
        """
        Retrieves the most recent samples from the model.

        Parameters:
        - num_samples: int - Number of recent samples to retrieve (default: 5, min: 1, max: 100)
        - token: str - Authorization token (inferred from permissions dependency)

        Returns:
        - JSONResponse: The most recent samples from the model
        """
        recent_samples = self.model.get_recent_samples(num_samples=num_samples)
        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=jsonable_encoder(recent_samples),
        )

    @udt_predict_metric.time()
    def predict(
        self,
        params: TokenAnalysisPredictParams,
        _=Depends(Permissions.verify_permission("read")),
    ):
        """
        Predicts the output based on the provided query parameters.

        Parameters:
        - text: str - The text for the sample to perform inference on
        - top_k: int - The number of results to return
        - data_type: str - The data type of the text. (unstructured or xml)
        - token: str - Authorization token (inferred from permissions dependency).

        Returns:
        - JSONResponse: Prediction results.

        Example Request Body:
        ```
        {
            "text": "What is artificial intelligence?",
            "top_k": 5,
            "data_type": "unstructured"
        }
        ```
        """
        start_time = time.perf_counter()

        text_length = len(params.text.split())
        udt_query_length.observe(text_length)

        results = self.model.predict(**params.model_dump())
        self.queries_ingested.log(1)
        self.queries_ingested_bytes.log(len(params.text))

        # TODO(pratik/geordie/yash): Add logging for search results text classification
        if isinstance(results, UnstructuredTokenClassificationResults):
            identified_count = len(
                [tags[0] for tags in results.predicted_tags if tags[0] != "O"]
            )
            self.tokens_identified.log(identified_count)
            self.logger.debug(
                f"Prediction complete with {identified_count} tokens identified",
                text_length=len(params.text),
            )

        elif isinstance(results, XMLTokenClassificationResults):
            self.tokens_identified.log(len(results.predictions))
            self.logger.debug(
                f"Prediction complete with {len(results.predictions)} predictions",
                text_length=len(params.text),
            )

        end_time = time.perf_counter()
        time_taken = end_time - start_time

        # Add time_taken to the response data
        response_data = {
            "prediction_results": jsonable_encoder(results),
            "time_taken": time_taken,
        }

        self.logger.debug(f"Prediction complete with time taken: {time_taken} seconds")

        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=response_data,
        )

    def new_report(
        self,
        documents: str = Form(...),
        tags: str = Form(None),
        files: List[UploadFile] = [],
        _=Depends(Permissions.verify_permission("write")),
    ):
        try:
            documents = DocumentList.model_validate_json(documents).documents
        except ValidationError as e:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="Invalid format for document report",
                data={"details": str(e), "documents": documents},
            )

        if not documents:
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message="No documents supplied for report. Must supply at least one document.",
            )

        report_id = str(uuid.uuid4())

        documents = download_local_files(
            files=files,
            file_infos=documents,
            dest_dir=self.reports_base_path / report_id / "documents",
        )

        with open(self.reports_base_path / report_id / "documents.json", "w") as file:
            json.dump([doc.model_dump() for doc in documents], file)

        custom_tags = tags if tags is not None else "[]"

        with self.get_session() as session:
            new_report = UDTReport(
                id=report_id,
                status="queued",
                submitted_at=datetime.utcnow(),
                updated_at=datetime.utcnow(),
                custom_tags=custom_tags,
            )
            session.add(new_report)
            session.commit()

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully submitted the documents to get the report, use the report_id to check the status.",
            data={"report_id": str(report_id)},
        )

    def get_report(
        self, report_id: str, _=Depends(Permissions.verify_permission("read"))
    ):
        with self.get_session() as session:
            report: UDTReport = session.query(UDTReport).get(report_id)

            if not report:
                return response(
                    status_code=status.HTTP_404_NOT_FOUND,
                    message=f"Report with ID '{report_id}' not found.",
                )

            with open(
                self.reports_base_path / report_id / "documents.json", "r"
            ) as file:
                documents = json.load(file)

            report_data = {
                "report_id": report.id,
                "status": report.status,
                "submitted_at": report.submitted_at,
                "updated_at": report.updated_at,
                "documents": documents,
                "msg": report.msg,
            }

            if report.status == "complete":
                report_file_path = (
                    self.reports_base_path / report_id / f"report_{report.attempt}.json"
                )

                if not report_file_path.exists():
                    if report.status == "complete":
                        return response(
                            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                            message=f"Processed reports directory for ID '{report_id}' is missing.",
                        )

                try:
                    with open(report_file_path) as file:
                        report_data["content"] = json.load(file)
                except Exception as e:
                    self.logger.error(
                        f"Failed to read document report {report_file_path}: {e}"
                    )

                    return response(
                        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                        message=f"failed to load report content",
                    )

            return response(
                status_code=status.HTTP_200_OK,
                message="Successfully retrieved the report details.",
                data=jsonable_encoder(report_data),
            )

    def delete_report(
        self, report_id: str, _=Depends(Permissions.verify_permission("write"))
    ):
        with self.get_session() as session:
            report = session.query(UDTReport).get(report_id)

            if not report:
                return response(
                    status_code=status.HTTP_404_NOT_FOUND,
                    message=f"Report with ID '{report_id}' not found.",
                )

            session.delete(report)
            session.commit()

        report_path = self.reports_base_path / report_id
        if report_path.exists() and report_path.is_dir():
            try:
                shutil.rmtree(report_path)
            except Exception as e:
                return response(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    message=f"Failed to delete report files for ID '{report_id}'.",
                    data={"details": str(e)},
                )

        return response(
            status_code=status.HTTP_200_OK,
            message=f"Successfully deleted report with ID '{report_id}'.",
        )

    def list_reports(self, _=Depends(Permissions.verify_permission("read"))):
        with self.get_session() as session:
            reports = session.query(UDTReport).all()

        data = [
            {
                "report_id": report.id,
                "status": report.status,
                "submitted_at": report.submitted_at,
                "updated_at": report.updated_at,
            }
            for report in reports
        ]

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully retrieved the reports.",
            data=jsonable_encoder(data),
        )

    def update_report_status(
        self,
        report_id: str,
        params: UpdateReportRequest,
        _=Depends(verify_job_token),
    ):
        try:
            with self.get_session() as session:
                report = session.query(UDTReport).get(report_id)
                if not report:
                    return response(
                        status_code=status.HTTP_404_BAD_REQUEST,
                        message=f"report {report_id} not found",
                    )

                if report.attempt != params.attempt:
                    return response(
                        status_code=status.HTTP_400_BAD_REQUEST,
                        message="invalid attempt number, this report has been assigned to a new worker",
                    )

                report.status = params.new_status
                report.updated_at = datetime.utcnow()
                report.msg = params.msg
                session.commit()
                self.logger.info(
                    f"updated status of report {report_id} to {params.new_status}"
                )
                return response(
                    status_code=status.HTTP_200_OK,
                    message="updated report status",
                )
        except Exception as e:
            self.logger.error(
                f"error updating report status of {report_id} to {params.new_status}: {e}"
            )
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="An error occurred while updating report status.",
                data={"details": str(e)},
            )

    def next_unprocessed_report(self, _=Depends(verify_job_token)):
        try:
            self.logger.info("checking for unprocessed reports")
            with self.get_session() as session:
                report = (
                    session.query(UDTReport)
                    .where(
                        or_(
                            UDTReport.status == "queued",
                            and_(
                                UDTReport.status == "in_progress",
                                UDTReport.attempt < MAX_ATTEMPTS,
                                UDTReport.updated_at
                                < (datetime.utcnow() - REPORT_TIMEOUT),
                            ),
                        )
                    )
                    .order_by(UDTReport.submitted_at.asc())  # Process oldest first
                    .with_for_update(skip_locked=True)  # Lock row to prevent conflicts
                    .first()
                )
                if report:
                    self.logger.info(f"found unprocessed report: {report.id}")
                    report.status = "in_progress"
                    report.attempt += 1
                    report.updated_at = datetime.utcnow()
                    session.commit()

                    return response(
                        status_code=status.HTTP_200_OK,
                        message="found unprocessed report",
                        data={"report_id": report.id, "attempt": report.attempt},
                    )
                else:
                    self.logger.info("no unprocessed reports found")

                    return response(
                        status_code=status.HTTP_200_OK,
                        message="no unprocessed reports found",
                        data={"report_id": None, "attempt": None},
                    )

        except Exception as e:
            self.logger.error(f"error checking for unprocessed report: {e}")
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="An error occurred while checking for unprocessed reports.",
                data={"details": str(e)},
            )
