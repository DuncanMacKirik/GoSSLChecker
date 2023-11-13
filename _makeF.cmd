@echo off
echo Building for FreeBSD...
del SSLChecker 2>nul
set GOARCH=amd64
set GOOS=freebsd
go get -v
go build -ldflags="-s -w" -v -x
upx --best SSLChecker