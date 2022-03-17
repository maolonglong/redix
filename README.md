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
- AUTH
- PING
- QUIT
- SHUTDOWN

## 安装

### 二进制

提前编译好的二进制文件，下载直接运行：[releases](https://github.com/MaoLongLong/redix/releases)

### 源码编译

```bash
$ go install go.chensl.me/redix/server/cmd/redix-server@main
...
$ redix-server # 数据默认存放在当前目录的 data 文件夹下
 _____          _ _
|  __ \        | (_)
| |__) |___  __| |___  __
|  _  // _ \/ _` | \ \/ /
| | \ \  __/ (_| | |>  <
|_|  \_\___|\__,_|_/_/\_\  commit=2f449de

{"level":"info","ts":1645346199.62673,"caller":"redix/redix.go:80","msg":"redix server started","host":"0.0.0.0","port":6380,"data_dir":"/Users/.../go/src/go.chensl.me/redix/data"}
```

使用 `redis-cli` 或 `iredis` 连接：

```bash
$ redis-cli -p 6380
```

### Docker

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
