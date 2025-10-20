#!/bin/sh

go test -coverprofile="coverage.out" -covermode count -tags test .
gocover-cobertura < coverage.out > coverage.xml