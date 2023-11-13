#!/bin/bash

echo Building for Linux...
rm SSLChecker &>/dev/null
go get -v
CGO_ENABLED=0 go build -ldflags="-s -w" -v -x
upx --best SSLChecker
