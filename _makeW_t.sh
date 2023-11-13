#!/bin/bash
echo Building tools for Windows...
export GOARCH=amd64
export GOOS=windows
go tool dist install -v pkg/runtime
go install -v -a std
