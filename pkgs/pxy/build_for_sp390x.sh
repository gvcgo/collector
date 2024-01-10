#!/bin/sh

export GOOS="linux"
export GOARCH="s390x"
go build -ldflags "-s -w" -o pxyc .
