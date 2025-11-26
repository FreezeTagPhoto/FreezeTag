.PHONY: run test mocks lint tidy clean
.SILENT:

TAGS=sqlite_math_functions,sqlite_foreign_keys

export CGO_CFLAGS_ALLOW=-Xpreprocessor
ifeq ($(OS),WINDOWS_NT)
# we might have to do Windows-specific stuff eventually
else 
	UNAME=$(shell uname -s)
	ifeq ($(UNAME),Darwin)
		CGO_LDFLAGS=-Wl,-no_warn_duplicate_libraries
	endif
endif
# ldflags to stop MacOS from complaining about a warning that doesn't matter constantly
export CGO_LDFLAGS
export CGO_ENABLED=1

tidy:
	go mod tidy

test: mocks
	mkdir -p coverage
	go test -coverprofile="coverage/coverage.out.tmp" -covermode count -tags=test,$(TAGS) ./... | grep -v "mocks"
	TESTPASSED=$$?
	cat coverage/coverage.out.tmp | grep -v "_mock.go" > coverage/coverage.out
	rm coverage/coverage.out.tmp
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	GOFLAGS="-tags=test,$(TAGS)" go tool gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml
	go tool cover -func coverage/coverage.out | tail -n 1
	exit $$TESTPASSED

lint: mocks
	golangci-lint run

run: tidy
	go run -tags=$(TAGS) ./cmd/server

mocks: tidy
	rm -rf ./mocks
	go tool mockery --config=./.mockery.yml

clean:
	rm -rf ./coverage
	rm -rf ./mocks