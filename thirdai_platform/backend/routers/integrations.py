from backend.auth_dependencies import verify_access_token, global_admin_only
from fastapi import APIRouter, Depends, HTTPException, status
from platform_common.utils import get_section, response
import pathlib

integrations_router = APIRouter()

root_folder = pathlib.Path(__file__).parent

#TODO(david) generate documentation?
docs_file = root_folder.joinpath("../../docs/integrations_endpoints.txt")

with open(docs_file) as f:
    docs = f.read()

# @integrations_router.get(
#     "/llm-provider-api-key",
#     summary="Get LLM Provider API Key",
#     description=get_section(docs, "Get LLM Provider API Key"),
#     dependencies=[Depends(verify_access_token)],
# )
# def get_llm_provider_api_key():
#     try:
#         # key key from vault
#         openai_api_key = "# openai key"
#         return {"openai_api_key": openai_api_key}
#     except Exception as e:
#         raise HTTPException(status_code=500, detail="Failed to retrieve OpenAI API key.")


# @integrations_router.post(
#     "/llm-provider-api-key",
#     summary="Set LLM Provider API Key",
#     description=get_section(docs, "Set LLM Provider API Key"),
#     dependencies=[Depends(global_admin_only)],
# )
# def set_llm_provider_api_key(api_key: str):
#     try:
#         # store key in vault
#         return {"message": "OpenAI API key stored successfully."}
#     except Exception as e:
#         raise HTTPException(status_code=500, detail="Failed to store OpenAI API key.")


# @integrations_router.delete(
#     "/llm-provider-api-key",
#     summary="Delete LLM Provider API Key",
#     description=get_section(docs, "Delete LLM Provider API Key"),
#     dependencies=[Depends(global_admin_only)],
# )
# def delete_llm_provider_api_key():
#     try:
#         # delete key from vault
#         return {"message": "OpenAI API key deleted successfully."}
#     except Exception as e:
#         raise HTTPException(status_code=500, detail="Failed to delete OpenAI API key.")


#TODO(david): support having multiple self-hosted LLM endpoints
@integrations_router.get(
    "/self-hosted-llm",
    summary="Get Self-Hosted LLM Integration",
    description=get_section(docs, "Get Self-Hosted LLM Integration"),
    dependencies=[Depends(verify_access_token)],
)
def get_self_hosted_llm():
    try:
        # store self hosted info in vault
        integration_details = {"endpoint": "# endpoint", "api_key": "# api key"}
        return integration_details
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to retrieve self-hosted LLM integration.")


@integrations_router.post(
    "/self-hosted-llm",
    summary="Store Self-Hosted LLM Integration",
    description=get_section(docs, "Store Self-Hosted LLM Integration"),
    dependencies=[Depends(global_admin_only)],
)
def set_self_hosted_llm(endpoint: str, api_key: str):
    try:
        # get self hosted info in vault
        # check if any models are currently configured with the self-hosted llm and give a warning or something
        
        return {"message": "Self-hosted LLM integration stored successfully."}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to store self-hosted LLM integration.")


@integrations_router.delete(
    "/self-hosted-llm",
    summary="Delete Self-Hosted LLM Integration",
    description=get_section(docs, "Delete Self-Hosted LLM Integration"),
    dependencies=[Depends(global_admin_only)],
)
def delete_self_hosted_llm():
    try:
        # delete self hosted info in vault
        # check if any models are currently configured with the self-hosted llm and give a warning or something
        return {"message": "Self-hosted LLM integration deleted successfully."}
    except Exception as e:
        raise HTTPException(status_code=500, detail="Failed to delete self-hosted LLM integration.")
