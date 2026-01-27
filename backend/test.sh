#!/bin/sh

set -e

# pass build tags to this script
TAGS=$1

mkdir -p coverage
go test -coverprofile="coverage/coverage.out" -covermode count -tags=test,$TAGS $(go list ./... | grep -v -e /mocks/ -e /cmd/)
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
GOFLAGS="-tags=test,$TAGS" go tool gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml
go tool cover -func coverage/coverage.out | tail -n 1