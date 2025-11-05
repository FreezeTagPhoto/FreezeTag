#!/bin/sh

mkdir -p coverage
go test -coverprofile="coverage/coverage.out" -covermode count -tags test ./...
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml