.PHONY: run test mocks lint tidy clean docs build deploy kill format
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
	./test.sh $(TAGS)

lint: mocks docs
	golangci-lint run

format:
	./format.sh

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
