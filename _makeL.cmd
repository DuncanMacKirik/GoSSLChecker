@echo off
echo Building for Linux...
del SSLChecker 2>nul
set GOARCH=amd64
set GOOS=linux
go build -ldflags="-s -w" -v -x
C:\Tools\UPX\upx --best SSLChecker