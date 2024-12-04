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

path = "./phone_guardrail.udt"
model.save(path)
