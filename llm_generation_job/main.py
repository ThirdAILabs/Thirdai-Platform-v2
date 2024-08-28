from typing import Any, Dict
import os
from urllib.parse import urljoin
import requests
from fastapi import FastAPI, WebSocket
from fastapi.middleware.cors import CORSMiddleware
from pydantic import ValidationError
from llms import DynamicLLM
from pydantic_models import GenerateArgs
import json

app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Load model configuration from JSON file
def load_model_config(config_file: str) -> Dict[str, Any]:
    with open(config_file, "r") as f:
        return json.load(f)

model_config = load_model_config("model_config.json")

# Initialize default keys
default_keys = {
    "openai": os.getenv("OPENAI_KEY", ""),
    "cohere": os.getenv("COHERE_KEY", ""),
}

# Register models dynamically based on the loaded configuration
model_registry = {}
for provider_name, config in model_config.items():
    model_registry[provider_name] = lambda config=config: DynamicLLM(config)

@app.websocket("/generate")
async def generate(websocket: WebSocket):
    await websocket.accept()
    while True:
        data = await websocket.receive_text()
        try:
            generate_args = GenerateArgs.parse_raw(data)
            break
        except ValidationError as e:
            await websocket.send_json(
                {
                    "status": "error",
                    "detail": "Invalid arguments",
                    "errors": e.errors(),
                    "end_of_stream": True,
                }
            )
            return
        except Exception as e:
            await websocket.send_json(
                {"status": "error", "detail": "Unexpected error", "end_of_stream": True}
            )
            return

    key = generate_args.key or default_keys.get(generate_args.provider.lower())
    if not key:
        await websocket.send_json(
            {
                "status": "error",
                "detail": "No generative AI key provided",
                "end_of_stream": True,
            }
        )
        return

    llm_class = model_registry.get(generate_args.provider.lower())
    if llm_class is None:
        await websocket.send_json(
            {
                "status": "error",
                "detail": "Unsupported provider",
                "end_of_stream": True,
            }
        )
        return

    llm = llm_class()

    generated_response = ""
    try:
        async for next_word in llm.stream(
            key=key, query=generate_args.query, model=generate_args.model
        ):
            generated_response += next_word
            await websocket.send_json(
                {"status": "success", "content": next_word, "end_of_stream": False}
            )
    except Exception as e:
        print("Error", e)
        await websocket.send_json(
            {
                "status": "error",
                "detail": "Error while generating content",
                "end_of_stream": True,
            }
        )
    await websocket.send_json(
        {"status": "success", "content": "", "end_of_stream": True}
    )

    if (
        generate_args.original_query is not None
        and generate_args.cache_access_token is not None
    ):
        try:
            res = requests.post(
                urljoin(os.environ["MODEL_BAZAAR_ENDPOINT"], "/cache/insert"),
                params={
                    "query": generate_args.original_query,
                    "llm_res": generated_response,
                },
                headers={
                    "Authorization": f"Bearer {generate_args.cache_access_token}",
                },
            )
            if res.status_code != 200:
                print(f"LLM Cache Insertion failed: {res}")
        except Exception as e:
            print("LLM Cache Insert Error", e)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="localhost", port=8000)
