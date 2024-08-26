import json
import logging
import time
import traceback
from typing import List

from pydantic import parse_obj_as
from pydantic_models.inputs import AssociateInputSingle, UpvoteInputSingle
from routers.model import get_model
from utils import Status, now

# Configure logging
logging.basicConfig(level=logging.INFO)


def process_task(task):
    """
    Process a single task based on the task action.

    Args:
        task (dict): The task data fetched from Redis.
    """
    model = get_model()
    task_id = task.get("task_id")
    try:
        # Update task status to "in_progress"
        task["status"] = Status.in_progress  # Use the enum value for status
        task["last_modified"] = now()  # Convert datetime to string
        model.redis_client.hset(
            f"task:{task_id}",
            mapping={
                k: json.dumps(v) if isinstance(v, (list, dict)) else v
                for k, v in task.items()
            },
        )

        action = task.get("action")
        model_id = task.get("model_id")

        if action == "upvote":
            # Deserialize and load back into Pydantic model
            text_id_pairs = parse_obj_as(
                List[UpvoteInputSingle], task.get("text_id_pairs", "[]")
            )
            model.upvote(text_id_pairs=text_id_pairs, token=task.get("token"))
            logging.info(
                f"Successfully upvoted for model_id: {model_id}, task_id: {task_id}"
            )

        elif action == "associate":
            # Deserialize and load back into Pydantic model
            text_pairs = parse_obj_as(
                List[AssociateInputSingle], task.get("text_pairs", "[]")
            )
            model.associate(text_pairs=text_pairs, token=task.get("token"))
            logging.info(
                f"Successfully associated text pairs for model_id: {model_id}, task_id: {task_id}"
            )

        elif action == "delete":
            # Deserialize and load back into Pydantic model
            source_ids = json.loads(task.get("source_ids", "[]"))
            model.delete(source_ids=source_ids, token=task.get("token"))
            logging.info(
                f"Successfully deleted sources for model_id: {model_id}, task_id: {task_id}"
            )

        elif action == "insert":
            documents = task.get("documents", "[]")  # Decode JSON
            model.insert(documents=documents, token=task.get("token"))
            logging.info(
                f"Successfully inserted documents for model_id: {model_id}, task_id: {task_id}"
            )

        # Mark task as completed
        task["status"] = Status.complete
        task["last_modified"] = now()
        model.redis_client.hset(
            f"task:{task_id}",
            mapping={
                k: json.dumps(v) if isinstance(v, (list, dict)) else v
                for k, v in task.items()
            },
        )

    except Exception as e:
        logging.error(f"Failed to process task {task_id}: {str(e)}")
        traceback.print_exc()
        # Update task status to "failed" and log the error
        task["status"] = Status.failed
        task["last_modified"] = now()
        task["message"] = str(e)
        model.redis_client.hset(
            f"task:{task_id}",
            mapping={
                k: json.dumps(v) if isinstance(v, (list, dict)) else v
                for k, v in task.items()
            },
        )
    finally:
        model.redis_client.srem(f"tasks_by_model:{model_id}", task_id)
        model.redis_client.delete(f"task:{task_id}")
        logging.info(f"Task {task_id} removed from Redis.")


def main():
    model = get_model()
    redis_client = model.redis_client
    model_id = model.general_variables.model_id

    while True:
        task_ids = redis_client.smembers(f"tasks_by_model:{model_id}")

        for task_id in task_ids:
            task_id = task_id.decode()  # Decode bytes to string
            task_data = redis_client.hgetall(f"task:{task_id}")
            if task_data:
                task_data = {k.decode(): v.decode() for k, v in task_data.items()}
                for key in ["text_id_pairs", "text_pairs", "documents", "source_ids"]:
                    if key in task_data:
                        task_data[key] = json.loads(task_data[key])

                # Process the task
                process_task(task_data)

        time.sleep(10)


if __name__ == "__main__":
    main()
