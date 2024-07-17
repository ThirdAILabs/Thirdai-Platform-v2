from abc import ABC


class Config(ABC):
    name = None

    base_path = "/Users/yashwanthadunukota/neuraldb-enterprise-services/headless/data"
    doc_type = "local"
    nfs_original_base_path = "/opt/neuraldb_enterprise/"
    unsupervised_paths = []
    supervised_paths = []
    test_paths = []
    insert_paths = []

    id_column = None
    strong_columns = None
    weak_columns = None
    reference_columns = None
    query_column = None
    id_delimiter = None

    model_cores = 2
    model_memory = 2000
    input_dim = 10000
    hidden_dim = 1024
    output_dim = 5000
    allocation_memory = 1000
    allocation_cores = 2

    epochs = 5

    retriever = "hybrid"


class Scifact(Config):
    name = "scifact"

    unsupervised_paths = [
        "scifact/unsupervised_1.csv",
        "scifact/unsupervised_2.csv",
    ]
    supervised_paths = [
        "scifact/supervised_1.csv",
        "scifact/supervised_2.csv",
    ]
    test_paths = ["scifact/test_1.csv", "scifact/test_2.csv"]
    insert_paths = ["scifact/insert.pdf"]

    strong_columns = ["TITLE"]
    weak_columns = ["TEXT"]
    reference_columns = ["TITLE", "TEXT"]
    id_column = "id"
    query_column = "query"
    id_delimiter = ":"
