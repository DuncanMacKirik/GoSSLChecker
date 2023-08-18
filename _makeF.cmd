@echo off
echo Building for FreeBSD...
del SSLChecker 2>nul
set GOARCH=amd64
set GOOS=freebsd
go build -ldflags="-s -w" -v -x
rem C:\Tools\UPX\upx --best SSLChecker