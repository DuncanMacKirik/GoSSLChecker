#!/bin/bash

echo Building for Windows...
rm SSLChecker.exe &>/dev/null
export GOARCH=amd64
export GOOS=windows
#export CGO_ENABLED=0
go get -v
go build -ldflags="-s -w" -v -x
upx --best SSLChecker.exe
