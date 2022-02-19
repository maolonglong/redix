FROM golang:1.17-alpine AS builder

ARG BRANCH=""
ARG COMMIT=""

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://proxy.golang.org

WORKDIR /go/src/go.chensl.me/redix
COPY . .

RUN go build -o /usr/bin/redix-server \
    -ldflags "-w -s -X main.branch=${BRANCH} -X main.commit=${COMMIT}" \
    ./cmd/redix-server

RUN apk add --no-cache upx && upx -9 /usr/bin/redix-server

FROM scratch

COPY --from=builder /usr/bin/redix-server /usr/bin/redix-server

EXPOSE 6380

ENTRYPOINT [ "redix-server" ]
