#!/bin/sh

mkdir -p coverage
go test -coverprofile="coverage/coverage.out" -covermode count -tags test ./...
gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml