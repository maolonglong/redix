#!/usr/bin/env bash

set -euxo pipefail

COMMIT=$(git rev-parse --short HEAD)

docker build --build-arg "COMMIT=${COMMIT}" -t maolonglong/redix .
