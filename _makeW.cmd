@echo off
echo Building for Windows...
del SSLChecker.exe 2>nul
go build -ldflags="-s -w" -v -x
C:\Tools\UPX\upx --best SSLChecker.exe