import glob
import json
from shutil import copytree
import os

from data_manager import schema
from data_manager.schema import get_session
from settings import get_settings

filenames = glob.glob(
    "/share/data/dataset_inventory/token_classification/**/*data.json"
)

# Adding entry to DB
with get_session() as session:
    for file in filenames:
        with open(file, "r") as f:
            data = json.load(f)
            entry = schema.Catalog(
                id=data["id"],
                name=data["name"] if data["name"] else "unnamed_data",
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
