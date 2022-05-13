FROM golang:1.18-bullseye AS build-env

ARG COMMIT=""

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://proxy.golang.org

WORKDIR /go/src/go.chensl.me/redix
COPY . .

RUN go build -o /go/bin/redix-server \
    -trimpath -buildvcs=false \
    -ldflags "-s -w -X main.commit=${COMMIT}" \
    ./server/cmd/redix-server

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends upx-ucl \
    && upx --best /go/bin/redix-server

FROM gcr.io/distroless/static
COPY --from=build-env /go/bin/redix-server /
ENTRYPOINT [ "/redix-server" ]
