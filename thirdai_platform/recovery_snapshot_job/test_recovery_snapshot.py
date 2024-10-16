import json
import os
import random
import shutil
import subprocess
import tempfile
from unittest.mock import MagicMock, patch

import pytest
from dotenv import load_dotenv
from platform_common.pydantic_models.recovery_snapshot import BackupConfig
from recovery_snapshot_job.run import perform_backup
from sqlalchemy import NullPool, create_engine, text

MODEL_BAZAAR_DIR = "./model_bazaar_tmp"


@pytest.fixture(scope="session")
def db_engine():
    """Fixture to create and drop the database."""
    load_dotenv()  # Load environment variables

    # Get the base DB URI
    db_uri = os.getenv("DB_BASE_URI")
    db_name = f"model_bazaar_{random.randint(0, 1e6)}"

    # Create the engine and database
    engine = create_engine(db_uri, isolation_level="AUTOCOMMIT", poolclass=NullPool)

    with engine.connect() as conn:
        conn.execute(text(f"CREATE DATABASE {db_name}"))

    # Provide the full DB URI with the test database
    os.environ["DATABASE_URI"] = f"{db_uri}/{db_name}"

    yield engine  # This will be passed into tests

    # Teardown: Drop the database after all tests are done
    with engine.connect() as conn:
        conn.execute(text(f"DROP DATABASE {db_name}"))


@pytest.fixture(autouse=True)
def setup_and_teardown(db_engine):
    """Fixture to setup temporary model_bazaar directory and config file."""
    # Setup: create a temporary directory for the model_bazaar_dir
    temp_dir = tempfile.mkdtemp()
    os.environ["MODEL_BAZAAR_DIR"] = temp_dir

    # Create a config.json file with local backup settings
    config_path = os.path.join(temp_dir, "backup_config.json")
    config = BackupConfig(
        cloud_provider=None,  # No cloud provider, so local backup
        interval_minutes=None,  # One-time backup
        backup_limit=2,  # Set backup limit to 2
    )
    with open(config_path, "w") as config_file:
        json.dump(config.dict(), config_file)

    # Set additional environment variables required for the backup
    os.environ["CONFIG_PATH"] = config_path

    yield  # This allows the test to run after the setup

    # Teardown: remove the temporary directory after the test completes
    shutil.rmtree(temp_dir)

    # Close any lingering sessions
    db_engine.dispose()


@patch("subprocess.run")  # Mock subprocess.run for pg_dump
def test_local_backup(mock_subprocess_run):
    """Test to check local backup functionality with mocked pg_dump."""
    config_path = os.getenv("CONFIG_PATH")

    # Simulate a successful pg_dump execution
    mock_subprocess_run.return_value = MagicMock(returncode=0)

    # Perform the backup immediately
    perform_backup(config_path)

    # Check if the backups directory is created in the model_bazaar_dir
    backups_dir = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "backups")
    assert os.path.exists(backups_dir), "Backups directory should be created"

    # Check if a backup zip file was created in the backups directory
    backup_files = [
        f
        for f in os.listdir(backups_dir)
        if f.startswith("backup_") and f.endswith(".zip")
    ]
    assert len(backup_files) > 0, "At least one backup file should be created"

    # Verify if the DB dump was created as part of the backup process (mocked)
    db_dump_path = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "db_backup.sql")
    assert os.path.exists(db_dump_path), "Database dump file should be created"

    # Ensure that pg_dump (subprocess.run) was called
    mock_subprocess_run.assert_called_once()


@patch("subprocess.run")  # Mock subprocess.run for pg_dump
def test_pg_dump_failure(mock_subprocess_run):
    """Test to check handling of pg_dump failure."""
    config_path = os.getenv("CONFIG_PATH")

    # Simulate a failure of pg_dump (non-zero exit code)
    mock_subprocess_run.side_effect = subprocess.CalledProcessError(
        returncode=1, cmd="pg_dump"
    )

    # Perform the backup and expect it to fail due to the mocked pg_dump failure
    with pytest.raises(Exception):
        perform_backup(config_path)

    # Ensure that pg_dump (subprocess.run) was called
    mock_subprocess_run.assert_called_once()


@patch("subprocess.run")  # Mock subprocess.run for pg_dump
def test_backup_limit(mock_subprocess_run):
    """Test to check if the backup limit is respected."""
    config_path = os.getenv("CONFIG_PATH")

    # Simulate a successful pg_dump execution
    mock_subprocess_run.return_value = MagicMock(returncode=0)

    # Perform the backup three times to exceed the backup limit of 2
    perform_backup(config_path)
    perform_backup(config_path)
    perform_backup(config_path)

    # Check if the backup directory has no more than 2 backups
    backups_dir = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "backups")
    backup_files = [
        f
        for f in os.listdir(backups_dir)
        if f.startswith("backup_") and f.endswith(".zip")
    ]
    assert (
        len(backup_files) == 2
    ), "Backup limit should enforce only 2 backups to be retained"


@pytest.mark.parametrize("backup_limit", [1, 3, 5])
@patch("subprocess.run")  # Mock subprocess.run for pg_dump
def test_parametrized_backup_limit(mock_subprocess_run, backup_limit):
    """Test to check parametrized backup limit behavior."""
    config_path = os.getenv("CONFIG_PATH")

    # Simulate a successful pg_dump execution
    mock_subprocess_run.return_value = MagicMock(returncode=0)

    # Load the existing config and modify the backup limit
    with open(config_path, "r") as config_file:
        config_data = json.load(config_file)

    config_data["backup_limit"] = backup_limit

    # Save the updated config
    with open(config_path, "w") as config_file:
        json.dump(config_data, config_file)

    # Run the backup multiple times to test the limit
    for _ in range(backup_limit + 2):  # Running more backups than the limit
        perform_backup(config_path)

    # Check if the backup directory has no more than the backup_limit
    backups_dir = os.path.join(os.getenv("MODEL_BAZAAR_DIR"), "backups")
    backup_files = [
        f
        for f in os.listdir(backups_dir)
        if f.startswith("backup_") and f.endswith(".zip")
    ]
    assert (
        len(backup_files) == backup_limit
    ), f"Backup limit should enforce only {backup_limit} backups to be retained"
