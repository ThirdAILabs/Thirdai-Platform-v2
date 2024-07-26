from fastapi import APIRouter, WebSocket
from llms import default_keys, model_classes
from pydantic import ValidationError
from pydantic_models import GenerateArgs

router = APIRouter()


@router.websocket("/generate")
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

    llm_class = model_classes.get(generate_args.provider.lower())
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

    try:
        async for next_word in llm.stream(
            key=key, query=generate_args.query, model=generate_args.model
        ):
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
