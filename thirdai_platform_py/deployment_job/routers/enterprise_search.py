from urllib.parse import urljoin

from deployment_job.permissions import Permissions
from fastapi import APIRouter, Depends, status
from fastapi.encoders import jsonable_encoder
from guardrail import Guardrail, LabelMap
from platform_common.logging import JobLogger, LogCode
from platform_common.pydantic_models.deployment import DeploymentConfig
from platform_common.utils import response
from prometheus_client import Summary
from pydantic_models import inputs
from reporter import Reporter
from requests import Session

query_metric = Summary("enterprise_search_query", "Enterprise Search Queries")


class EnterpriseSearchRouter:
    def __init__(self, config: DeploymentConfig, reporter: Reporter, logger: JobLogger):
        self.config = config
        self.logger = logger

        self.session = Session()
        self.retrieval_endpoint = urljoin(
            self.config.model_bazaar_endpoint,
            self.config.options["retrieval_id"] + "/",
        )
        self.logger.info(
            f"Retrieval endpoint set to {self.retrieval_endpoint}",
            code=LogCode.MODEL_INFO,
        )

        if self.config.options.get("guardrail_id", None):
            self.guardrail = Guardrail(
                guardrail_model_id=self.config.options["guardrail_id"],
                model_bazaar_endpoint=self.config.model_bazaar_endpoint,
                logger=self.logger,
            )
            self.logger.info(
                f"Guardrail initialized with ID {self.config.options['guardrail_id']}",
                code=LogCode.GUARDRAILS,
            )
        else:
            self.guardrail = None
            self.logger.info(
                "No guardrail configuration found for this model",
                code=LogCode.GUARDRAILS,
            )

        self.router = APIRouter()
        self.router.add_api_route("/search", self.search, methods=["POST"])
        self.router.add_api_route("/unredact", self.unredact, methods=["POST"])

    @query_metric.time()
    def search(
        self,
        params: inputs.NDBSearchParams,
        token_scheme=Depends(Permissions.verify_permission("read")),
    ):

        token, scheme = token_scheme

        if scheme == "api_key":
            headers = {"X-API-Key": token}
        else:
            headers = {"Authorization": f"Bearer {token}"}

        try:
            res = self.session.post(
                url=urljoin(self.retrieval_endpoint, "search"),
                json=params.model_dump(),
                headers=headers,
            )
        except Exception as e:
            self.logger.error(f"Exception during POST request: {e}")
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="Failed to make POST request: " + str(e),
            )

        if res.status_code != status.HTTP_200_OK:
            self.logger.error(
                f"Failed retrieval request with status code {res.status_code}. Response: {res.text}",
                code=LogCode.MODEL_PREDICT,
            )
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="Unable to get results from retrieval model: " + str(res),
            )

        try:
            results = inputs.EnterpriseSearchResults.model_validate(res.json()["data"])
        except Exception as e:
            self.logger.error(f"Exception during result parsing: {e}")
            return response(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                message="Error parsing results: " + str(e),
            )

        if self.guardrail:
            label_map = LabelMap()

            try:
                results.query_text = self.guardrail.redact_pii(
                    text=results.query_text,
                    label_map=label_map,
                    access_token=token,
                    auth_scheme=scheme,
                )

                for ref in results.references:
                    ref.text = self.guardrail.redact_pii(
                        text=ref.text,
                        label_map=label_map,
                        access_token=token,
                        auth_scheme=scheme,
                    )

                results.pii_entities = label_map.get_entities()
                self.logger.debug("Redacted PII from search results")
            except Exception as e:
                self.logger.error(f"Exception during PII redaction: {e}")

        return response(
            status_code=status.HTTP_200_OK,
            message="Successful",
            data=jsonable_encoder(results),
        )

    def unredact(
        self,
        args: inputs.UnredactArgs,
        _=Depends(Permissions.verify_permission("read")),
    ):
        if self.guardrail:
            unredacted_text = self.guardrail.unredact_pii(args.text, args.pii_entities)
            self.logger.debug("Unredacted text successfully")
            return response(
                status_code=status.HTTP_200_OK,
                message="Successful",
                data={"unredacted_text": unredacted_text},
            )
        else:
            message = "Cannot unredact text since this model was not configured with guardrails."
            self.logger.error(message, code=LogCode.GUARDRAILS)
            return response(
                status_code=status.HTTP_400_BAD_REQUEST,
                message=message,
            )
