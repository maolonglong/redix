#!/usr/bin/env bash

set -euxo pipefail

if [ ! -d "bin" ]; then
    mkdir bin
fi

COMMIT=$(git rev-parse --short HEAD)

go build -o bin/redix-server \
    -ldflags="-w -s -X main.commit=${COMMIT}" \
    ./cmd/redix-server
