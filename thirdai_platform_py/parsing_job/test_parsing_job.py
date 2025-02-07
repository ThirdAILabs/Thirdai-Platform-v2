import json
import os
import shutil
import uuid
from pathlib import Path

import pytest
from fastapi.testclient import TestClient

pytestmark = [pytest.mark.unit]

UPLOAD_ID = str(uuid.uuid4())


@pytest.fixture(scope="function")
def temp_model_bazaar_dir():
    path = "./tmp"
    os.mkdir(path)
    os.makedirs(os.path.join(path, "uploads", UPLOAD_ID))
    yield path
    shutil.rmtree(path)


def doc_dir():
    return os.path.join(os.path.dirname(__file__), "../train_job/sample_docs")


def upload(model_bazaar_dir, source_file):
    shutil.copyfile(
        source_file,
        os.path.join(model_bazaar_dir, "uploads", UPLOAD_ID, Path(source_file).name),
    )


def try_parse(client, upload_id, filename, metadata=None, options={}):
    return client.post(
        "/parse",
        json={
            "upload_id": upload_id,
            "filename": filename,
            "metadata": metadata,
            "options": options,
        },
    )


def parse(client, upload_id, filename, metadata=None, options={}):
    response = try_parse(client, upload_id, filename, metadata, options)
    print(response.text)
    assert response.status_code == 200


@pytest.mark.parametrize("document", ["articles.csv", "mutual_nda.pdf"])
def test_valid_parse(temp_model_bazaar_dir, document):
    os.environ["MODEL_BAZAAR_DIR"] = temp_model_bazaar_dir

    import parsing_job.main

    client = TestClient(parsing_job.main.app)

    upload(temp_model_bazaar_dir, os.path.join(doc_dir(), document))

    assert os.path.exists(
        os.path.join(temp_model_bazaar_dir, "uploads", UPLOAD_ID, Path(document).name)
    )

    parse(client, UPLOAD_ID, Path(document).name, metadata={"something": 42})

    json_file = os.path.join(
        temp_model_bazaar_dir,
        "uploads",
        UPLOAD_ID,
        "parsed",
        f"{Path(document).stem}.json",
    )

    with open(json_file, "r") as f:
        data = json.load(f)
        assert "document" in data.keys()
        assert "text" in data.keys()
        assert "metadata" in data.keys()
        assert len(data["text"]) == len(data["metadata"]) != 0
        assert data["metadata"][0]["something"] == 42


def test_missing_file(temp_model_bazaar_dir):
    os.environ["MODEL_BAZAAR_DIR"] = temp_model_bazaar_dir
    import parsing_job.main

    client = TestClient(parsing_job.main.app)

    missing_document = "non_existent_file.pdf"
    response = try_parse(client, UPLOAD_ID, missing_document)
    assert response.status_code == 404
    assert response.json()["detail"].endswith("does not exist")


def test_invalid_extension(temp_model_bazaar_dir):
    os.environ["MODEL_BAZAAR_DIR"] = temp_model_bazaar_dir
    import parsing_job.main

    client = TestClient(parsing_job.main.app)

    valid_document = "articles.csv"
    source_file = os.path.join(doc_dir(), valid_document)
    shutil.copyfile(
        source_file,
        os.path.join(temp_model_bazaar_dir, "uploads", UPLOAD_ID, "articles.something"),
    )

    response = try_parse(client, UPLOAD_ID, "articles.something")
    assert response.status_code == 404
    assert (
        response.json()["detail"]
        == "Invalid extension for doc, please use one of .pdf, .csv, .docx, or .html"
    )
