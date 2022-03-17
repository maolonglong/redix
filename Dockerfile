FROM golang:1.18-alpine AS builder

ARG COMMIT=""

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://proxy.golang.org

WORKDIR /code
COPY . .

RUN go build -o /usr/bin/redix-server \
    -trimpath -buildvcs=false \
    -ldflags "-w -s -X main.commit=${COMMIT}" \
    ./redix/cmd/redix-server

RUN apk add --no-cache upx && upx -9 /usr/bin/redix-server

FROM scratch

COPY --from=builder /usr/bin/redix-server /usr/bin/redix-server

EXPOSE 6380

ENTRYPOINT [ "redix-server" ]
