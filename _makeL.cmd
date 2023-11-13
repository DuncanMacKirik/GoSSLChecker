@echo off
echo Building for Linux...
del SSLChecker 2>nul
set GOARCH=amd64
set GOOS=linux
go get -v
go build -ldflags="-s -w" -v -x
upx --best SSLChecker