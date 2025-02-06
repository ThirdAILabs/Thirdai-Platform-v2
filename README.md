# ThirdAI-Platform

## Starting the Backend

To run model bazaar locally follow the local setup instructions under `local_setup/`. Then navigate to `thirdai_platform/` and follow the instructions in `thirdai_platform/cmd/model_bazaar`.

## Running Tests

### Unit Tests

To run the `model_bazaar` unit tests you can navigate to the `thirdai_platform/` directory and run `go test ./model_bazaar/tests`. To run a specific test pass the flag `--run <name of test>`, and to display output from tests (logs + print statements) pass the `-v` flag. 

### Integration Tests

To run the integration tests you can navigate to the `thirdai_platform/` directory and run `go test ./integration_tests`. To run a specific test pass the flag `--run <name of test>`, and to display output from tests (logs + print statements) pass the `-v` flag. 

## DB Migrations

To run db migrations navigate to `thirdai_platform/` and follow the instructions in `thirdai_platform/cmd/migration`.

## Documentation

Documentation is available in the `docs/` directory.

## Code Structure

From `thirdai_platform/`, the code is organized as follows:

- `/client`
    - This is a general client for interacting with the backend and deployments, it is written in go and used in the integration tests
- `/cmd`
    - `/model_bazaar`
        - Entrypoint for starting model bazaar
    - `/migrations`
        - Code for running db migrations.
- `/model_bazaar`
    - `/auth`
        - Keycloak logic, as well as general jwt managment
        - Permissions check for models
        - Has auth middleware that can verify users are authenticated for endpoints and check permissions for models
    - `/config`
        - Configs for train/deploy jobs
    - `/jobs`
        - code to start various jobs, frontend, llm dispatch/cache, on prem llm, telemetry, etc.
    - `/licensing`
        - Code to check licenses (uses the same format as the current platform licenses)
    - `/nomad`
        - This contains a client for interacting with nomad
        - Http nomad client
        - Job templates (using go template engine)
            - Structs that contain the fields required by each template
    - `/schema`
        - Database schema
        - utils for common queries like getting models, users, etc.
    - `/services`
        - Implementation of rest endpoints
    - `/storage`
        - Interface for interacting with shared storage (read/write/list/etc files)
        - This is so that if we want to change the storage engine to something like s3 there is a simple interface we can implement and swap in the code
            - or if we want to add something like encryption of files
    - `/tests`
        - Unit tests for the backend
        - This uses a nomad client stub and sqlite for the database
        - Requires no external setup to run (entirely self contained)
        - Utilities to create new env for each test, (fresh database, directory, etc.) so that tests are idempotent
- `/integration_tests`
    - General integration tests between backend and train/deployment jobs
    - Tests various training/deployment scenarios and other workflows

## Libraries Used (Go)

We are using these 3rd party libraries (all of which are fairly common and have thousands of stars on GitHub) 

- chi (http router and various utilities, extends standard library)
    - https://github.com/go-chi/chi
    - Also using jwt extension
        - https://github.com/go-chi/jwtauth
- gorm (sql orm library)
    - https://github.com/go-gorm/gorm
    - Also using the form sql drivers
- gormigrate(for db migrations)
    - https://github.com/go-gormigrate/gormigrate
    - Gorm has support for the necessary migration functions natively (add/drop column/table/index, etc.)
    - This library is very minimal, it adds versioning will still relying on the native gorm migration methods.
- google uuid (for generating uuids)
    - https://github.com/google/uuid
- go cloak (go Keycloak client)
    - [https://github.com/Nerzal/gocloakttps://github.com/Nerzal/gocloak](https://github.com/Nerzal/gocloak)
- go yaml (for parsing configs for telemetry setup)
    - https://github.com/go-yaml/yaml
- godotenv (for loading .env files, this is so that the local setup is compatible with our current backend)
    - https://github.com/joho/godotenv