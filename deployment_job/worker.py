import json
import logging
import time
import traceback
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
        model.redis_client.hset(f"task:{task_id}", mapping={k: json.dumps(v) if isinstance(v, (list, dict)) else v for k, v in task.items()})

        action = task.get("action")
        model_id = task.get("model_id")

        if action == "upvote":
            text_id_pairs = json.loads(task.get("text_id_pairs", "[]"))  # Decode JSON
            model.upvote(text_id_pairs=text_id_pairs, token=task.get("token"))
            logging.info(f"Successfully upvoted for model_id: {model_id}, task_id: {task_id}")

        elif action == "associate":
            text_pairs = json.loads(task.get("text_pairs", "[]"))  # Decode JSON
            model.associate(text_pairs=text_pairs, token=task.get("token"))
            logging.info(f"Successfully associated for model_id: {model_id}, task_id: {task_id}")

        elif action == "delete":
            source_ids = json.loads(task.get("source_ids", "[]"))  # Decode JSON
            model.delete(source_ids=source_ids, token=task.get("token"))
            logging.info(f"Successfully deleted sources for model_id: {model_id}, task_id: {task_id}")

        elif action == "insert":
            documents = json.loads(task.get("documents", "[]"))  # Decode JSON
            model.insert(documents=documents, token=task.get("token"))
            logging.info(f"Successfully inserted documents for model_id: {model_id}, task_id: {task_id}")

        # Mark task as completed
        task["status"] = Status.complete
        task["last_modified"] = now()
        model.redis_client.hset(f"task:{task_id}", mapping={k: json.dumps(v) if isinstance(v, (list, dict)) else v for k, v in task.items()})

    except Exception as e:
        logging.error(f"Failed to process task {task_id}: {str(e)}")
        traceback.print_exc()
        # Update task status to "failed" and log the error
        task["status"] = Status.failed
        task["last_modified"] = now()
        task["message"] = str(e)
        model.redis_client.hset(f"task:{task_id}", mapping={k: json.dumps(v) if isinstance(v, (list, dict)) else v for k, v in task.items()})


def list_all_redis_keys():
    model = get_model()
    redis_client = model.redis_client
    all_keys = redis_client.keys("*")
    print(all_keys, flush=True)
    # for key in all_keys:
    #     print(f"Key: {key.decode()}")
    #     if redis_client.type(key) == b'hash':
    #         key_data = redis_client.hgetall(key)
    #         decoded_data = {k.decode(): v.decode() for k, v in key_data.items()}
    #         print(decoded_data)
    #     elif redis_client.type(key) == b'set':
    #         members = redis_client.smembers(key)
    #         print({member.decode() for member in members})


def main():
    model = get_model()
    redis_client = model.redis_client
    model_id = model.general_variables.model_id

    while True:
        # Fetch task IDs specific to the current model_id
        task_ids = redis_client.smembers(f"tasks_by_model:{model_id}")
        print(task_ids)

        for task_id in task_ids:
            task_id = task_id.decode()  # Decode bytes to string
            # Retrieve and deserialize the task data
            task_data = redis_client.hgetall(f"task:{task_id}")
            if task_data:
                # Decode and deserialize fields as needed
                task_data = {k.decode(): v.decode() for k, v in task_data.items()}
                for key in ["text_id_pairs", "text_pairs", "documents", "source_ids"]:
                    if key in task_data:
                        task_data[key] = json.loads(task_data[key])

                # Process the task
                process_task(task_data)

        time.sleep(10)  # Adjust the sleep time according to your needs
        list_all_redis_keys()


if __name__ == "__main__":
    main()
