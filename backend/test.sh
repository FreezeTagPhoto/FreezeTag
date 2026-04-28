#!/bin/sh

set -e

# pass build tags to this script
TAGS=$1

mkdir -p coverage
go test -coverprofile="coverage/coverage.out" -covermode count -tags=test,$TAGS $(go list ./... | grep -v -e /mocks/ -e /cmd/)
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
GOFLAGS="-tags=test,$TAGS" go tool gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml
coverage=$(go tool cover -func coverage/coverage.out | tail -n 1 | awk '{print $3}')
echo "overall coverage: $coverage"
required=80.0%
if [ "$(echo "${required%\%} > ${coverage%\%}" | bc)" -eq 1 ]; then
    echo "coverage does not meet requirements"
    exit 1
fi