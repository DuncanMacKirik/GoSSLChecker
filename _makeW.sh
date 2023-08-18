#!/bin/bash

echo Building for Windows...
rm SSLChecker.exe &>/dev/null
export GOARCH=amd64
export GOOS=windows
#export CGO_ENABLED=0
go build -ldflags="-s -w" -v -x
./upx --brute SSLChecker.exe
