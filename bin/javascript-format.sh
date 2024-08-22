#!/bin/bash

BASEDIR=$(dirname "$0")
cd $BASEDIR/../frontend
pnpm exec prettier . --write 