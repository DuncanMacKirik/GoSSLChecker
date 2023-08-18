#!/bin/bash
echo Building tools for FreeBSD...
export GOARCH=amd64
export GOOS=freebsd
go tool dist install -v pkg/runtime
go install -v -a std

