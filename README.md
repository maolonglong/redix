# Redix

[![CircleCI](https://circleci.com/gh/MaoLongLong/redix/tree/main.svg?style=svg)](https://circleci.com/gh/MaoLongLong/redix/tree/main)

可能是我毕业设计的一小部分，一个精简版 Redis，它只是个“小玩具”

支持的命令：

- SET
- SETNX
- GET
- INCR
- INCRBY
- DECR
- DECRBY
- KEYS
- TTL
- EXPIRE
- DEL
- FLUSHALL
- FLUSHDB: 没有区分 db，所以直接调用的 FLUSHALL

## 从源码安装

```bash
$ go install go.chensl.me/redix/cmd/redix-server@latest
go: downloading go.chensl.me/redix v0.0.0-20220219084604-d5ac6cfcea68
$ redix-server # 数据默认存放在当前目录的 data 文件夹下
 _____          _ _
|  __ \        | (_)
| |__) |___  __| |___  __
|  _  // _ \/ _` | \ \/ /
| | \ \  __/ (_| | |>  <
|_|  \_\___|\__,_|_/_/\_\  branch= commit=

{"level":"info","ts":1645260845.3469691,"caller":"redix@v0.0.0-20220219084604-d5ac6cfcea68/redix.go:67","msg":"redix server listening","addr":"tcp://0.0.0.0:6380"}
```

使用 `redis-cli` 或 `iredis` 连接：

```bash
$ redis-cli -p 6380
```

## Docker

```bash
$ mkdir data
$ docker run -d -p 6380:6380 \
    --name redix \
    -v $PWD/data:/data \
    -e REDIX_DATA_DIR=/data \
    -e REDIX_PASSWORD=123456 \
    maolonglong/redix:latest
```

## 配置

可通过配置文件或环境变量进行配置（环境变量优先级更高），配置文件示例：[config.yml](./configs/config.yml)

配置文件可放置在（按顺序查找）：

- 当前目录 `config.yml`
- 用户目录 `.redix.yml`
- `/etc/redix/config.yml`

可用配置和默认值：

- REDIX_HOST: 0.0.0.0
- REDIX_PORT: 6380
- REDIX_PASSWORD: ""
- REDIX_DATA_DIR: ./data
