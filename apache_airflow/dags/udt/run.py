from  thirdai import neural_db as ndb
import thirdai


def run():
    thirdai.licensing.activate("002099-64C584-3E02C8-7E51A0-DE65D9-V3")
    db = ndb.NeuralDB(
        retriever="finetunable_retriever",
        on_disk=True,
    )

    doc = ndb.PDF("/Users/yashwanthadunukota/neuraldb-enterprise-services/headless/data/scifact/insert.pdf", save_extra_info=False, on_disk=True)

    db.insert([doc])
