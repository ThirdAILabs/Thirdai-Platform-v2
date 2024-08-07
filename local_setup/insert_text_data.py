import json
import os
from shutil import copytree

from data_manager import schema
from data_manager.schema import get_session
from settings import get_settings

root = "/share/data/dataset_inventory/text_classification"

# Adding entry to DB
with get_session() as session:
    for dirEntry in os.scandir(root):
        if not dirEntry.is_file():
            with open(os.path.join(root, dirEntry.name, "catalog.json"), "r") as f:
                data = json.load(f)
                entry = schema.Catalog(
                    id=data["id"],
                    name=data["name"],
                    task=data["task"],
                    sub_tasks=data["sub_tasks"],
                    input_feature=data["input_feature"],
                    target_feature=data["target_feature"],
                    target_labels=data["target_labels"],
                    splits=data["splits"],
                    generated=data["generated"],
                    num_generated_samples=data["num_generated_samples"],
                    user_inserted=False,
                )
                session.add(entry)
    session.commit()

# Copying files
copytree(
    "/share/data/dataset_inventory/token_classification",
    os.getenv("SHARE_DIR"),
    dirs_exist_ok=True,
)
