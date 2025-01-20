import json
import os
import shutil

from thirdai import bolt

model = bolt.UniversalDeepTransformer(
    data_types={
        "source": bolt.types.text(),
        "target": bolt.types.token_tags(tags=[], default_tag="O"),
    },
    target="target",
    rules=True,
    embedding_dimension=10,
)
model.add_ner_rule("PHONENUMBER")

model_dir = "./phone_guardrail"
os.makedirs(model_dir, exist_ok=True)

model.save(os.path.join(model_dir, "model.udt"))

with open(os.path.join(model_dir, "metadata.json"), "w") as f:
    json.dump({"Type": "nlp-token", "Attributes": {}}, f)

shutil.make_archive(model_dir, "zip", root_dir=model_dir)

shutil.rmtree(model_dir)
