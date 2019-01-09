
GOPATH ?= /usr/share/gocode
GOTEST_ARGS ?= 
CONFIG ?= config.yaml.test

build:
	GOPATH=$(GOPATH) go build kask.go config.go http.go logging.go storage.go

functional-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=functional -config $(CONFIG)

unit-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=unit

integration-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=integration -config $(CONFIG)

test: unit-test functional-test

clean:
	rm -f kask

.PHONY: build functional-test unit-test integration-test test clean
