#!/usr/bin/env bash

# 创建目录
mkdir -p ./data/rocketmq/namesrv/logs
mkdir -p ./data/rocketmq/namesrv/store
mkdir -p ./data/rocketmq/broker/logs
mkdir -p ./data/rocketmq/broker/store

# 设置目录权限
chmod -R 777 ./data/rocketmq/namesrv/logs
chmod -R 777 ./data/rocketmq/namesrv/store
chmod -R 777 ./data/rocketmq/broker/logs
chmod -R 777 ./data/rocketmq/broker/store

# 下载并启动容器，且为 后台 自动启动
docker-compose up -d


# 显示 rocketmq 容器
docker ps |grep rocketmq