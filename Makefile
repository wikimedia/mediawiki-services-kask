
GOPATH ?= /usr/share/gocode
GOTEST_ARGS ?= 
CONFIG ?= config.yaml.test
GO_PACKAGES := ./...

build:
	GOPATH=$(GOPATH) go build kask.go config.go http.go logging.go storage.go

functional-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=functional -config $(CONFIG)

unit-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=unit

integration-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=integration -config $(CONFIG)

deps:
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

check: deps
	@if [ -n "`goimports -l *.go`" ]; then \
        echo "goimports: format errors detected" >&2; \
        false; \
    fi
	@if [ -n "`gofmt -l *.go`" ]; then \
        echo "gofmt: format errors detected" >&2; \
        false; \
    fi
	golint -set_exit_status $(GO_PACKAGES)
	go vet -composites=false $(GO_PACKAGES)

test: unit-test functional-test check

clean:
	rm -f kask

.PHONY: build functional-test unit-test integration-test test clean
