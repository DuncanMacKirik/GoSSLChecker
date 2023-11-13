#!/bin/bash

echo Building for FreeBSD...
rm SSLChecker &>/dev/null
export GOARCH=amd64
export GOOS=freebsd
export CGO_ENABLED=0
go get -v
go build -ldflags="-s -w" -v -x
upx --best SSLChecker
