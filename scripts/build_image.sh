#!/usr/bin/env bash

set -euxo pipefail

BRANCH=$(git rev-parse --abbrev-ref HEAD)
COMMIT=$(git rev-parse --short HEAD)

docker build --build-arg "BRANCH=${BRANCH}" \
    --build-arg "COMMIT=${COMMIT}" \
    -t maolonglong/redix .
