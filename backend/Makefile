.PHONY: run test mocks lint tidy clean docs build deploy kill
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

docs: tidy
	go tool swag init -q -g cmd/server/main.go -o cmd/docs

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

lint: mocks docs
	golangci-lint run

run: docs
	go run -tags=$(TAGS) ./cmd/server

mocks: tidy
	rm -rf ./mocks
	go tool mockery --config=./.mockery.yml

clean:
	rm -rf ./coverage
	rm -rf ./mocks

build:
	docker build -t freezetag:backend .

deploy: build
	docker run -d --name freezetag_backend -p 3824:3824 freezetag:backend

kill:
	docker kill freezetag_backend
	docker rm freezetag_backend
