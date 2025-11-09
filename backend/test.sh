#!/bin/sh

TAGS='test,sqlite_math_functions,sqlite_foreign_keys'

mkdir -p coverage
go test -coverprofile="coverage/coverage.out" -covermode count -tags=$TAGS ./...
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
GOFLAGS="-tags=$TAGS" gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml
go tool cover -func coverage/coverage.out | tail -n 1