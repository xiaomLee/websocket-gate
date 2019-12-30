#!/usr/bin/env bash

echo "go env GO111MODULE=on"
go env GO111MODULE=on
echo "go env GOPROXY=https://goproxy.cn,direct"
go env GOPROXY=https://goproxy.cn,direct

echo "go build -o websocket-gate main.go"
go build -o websocket-gate main.go