#!/usr/bin/env bash

set -euxo pipefail

if [ ! -d "bin" ]; then
    mkdir bin
fi

COMMIT=$(git rev-parse --short HEAD)

CGO_ENABLED=0 go build -v -o bin/redix-server \
    -trimpath -buildvcs=false \
    -ldflags="-w -s -X main.commit=${COMMIT}" \
    ./server/cmd/redix-server
