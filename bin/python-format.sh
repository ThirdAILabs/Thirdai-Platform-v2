#!/bin/bash

BASEDIR=$(dirname "$0")
cd $BASEDIR/..
# Remove unused imports with autoflake
autoflake --remove-all-unused-imports --in-place --recursive .
black .
isort . --profile black