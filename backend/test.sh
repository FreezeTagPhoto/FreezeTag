#!/bin/sh

set -e

# pass build tags to this script
TAGS=$1

mkdir -p coverage
go test -coverprofile="coverage/coverage.out.tmp" -covermode count -tags=test,$TAGS $(go list ./... | grep -v /mocks/)
cat coverage/coverage.out.tmp | grep -v "_mock.go" > coverage/coverage.out
rm coverage/coverage.out.tmp
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
GOFLAGS="-tags=test,$TAGS" go tool gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml
go tool cover -func coverage/coverage.out | tail -n 1