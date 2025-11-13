.PHONY: run test
.SILENT:

TAGS=sqlite_math_functions,sqlite_foreign_keys

export CGO_FLAGS_ALLOW='-Xpreprocessor'
export CGO_ENABLED=1

test: mocks
	mkdir -p coverage
	go test -coverprofile="coverage/coverage.out" -covermode count -tags=test,$(TAGS) ./...
	TESTPASSED=$$?
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	GOFLAGS="-tags=test,$(TAGS)" gocover-cobertura < coverage/coverage.out > coverage/cobertura-coverage.xml
	go tool cover -func coverage/coverage.out | tail -n 1
	exit $$TESTPASSED

run:
	go run -tags=$(TAGS) .

mocks: 
	mockery --config=./.mockery.yml

clean:
	rm -rf ./coverage
	rm -rf ./mockery