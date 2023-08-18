#!/bin/bash

echo Building for FreeBSD...
rm SSLChecker &>/dev/null
export GOARCH=amd64
export GOOS=freebsd
export CGO_ENABLED=0
go build -ldflags="-s -w" -v -x
#./upx --brute SSLChecker
