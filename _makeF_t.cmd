@echo off
echo Building tools for FreeBSD...
set GOARCH=amd64
set GOOS=freebsd
go tool dist install -v pkg/runtime
go install -v -a std
