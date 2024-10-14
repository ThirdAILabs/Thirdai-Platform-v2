#!/bin/bash

mkdir -p ./registry_data

docker run --rm \
  -e admin_email=$1 \
  -e admin_password=$2 \
  -p 8080:8080 \
  --mount type=bind,source=./registry_data,target=/app/data \
  model_registry