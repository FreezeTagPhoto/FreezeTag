#!/bin/sh

mkdir -p coverage
go test -coverprofile="coverage/coverage.out" -covermode count -tags test ./...
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
go tool cover -func coverage/coverage.out | tail -n 1
gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml