@echo off
echo Building for Windows...
del SSLChecker.exe 2>nul
go get -v
go build -ldflags="-s -w" -v -x
upx --best SSLChecker.exe